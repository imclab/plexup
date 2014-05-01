package main

import (
	"io"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

const loggingTag = "plexup"
const plexupPort = 25010

type plexup struct {
	running chan bool
}

func pmsController(c chan bool, logger *syslog.Writer) {
	finished := make(chan struct{}, 1)
	finished <- struct{}{}
	for {
		logger.Notice("controller: czekam na desiredState")
		desiredState := <-c
		logger.Notice("controller: dostałem nowy desiredState z kanału")
		switch desiredState {
		case true:
			select {
			default:
				logger.Notice("controller: I think pms is already running")
			case <-finished:
				logger.Notice("Turning Plex Media Server on.")
				cmd := exec.Command("/usr/bin/caffeinate", "/Applications/Plex Media Server.app/Contents/MacOS/Plex Media Server")
				go func() {
					cmd.Run()
					finished <- struct{}{}
				}()
				// TODO: hit http://127.0.0.1:32400/library/sections/all/refresh
			}
		case false:
			logger.Notice("Turning Plex Media Server off.")
			gpid, _ := syscall.Getpgid(os.Getpid())
			// TODO: avoid calling pkill
			exec.Command("pkill", "-g", strconv.Itoa(gpid), "Plex Media Server").Run()
		}
	}
}

func (pms *plexup) on(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "on handler\n")
	pms.running <- true
}

func (pms *plexup) off(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "off handler\n")
	pms.running <- false
}

func deathRattler(c chan os.Signal, logger *syslog.Writer) {
	<-c
	logger.Notice("SIGTERM received - killing all processes in my process group.")
	// Kill whole process group.
	syscall.Kill(0, syscall.SIGTERM)
	os.Exit(1)
}

func main() {
	var address = ":" + strconv.Itoa(plexupPort)
	var err error
	logger, err := syslog.New(syslog.LOG_USER|syslog.LOG_NOTICE, loggingTag)
	if err != nil {
		log.Fatalln(err)
	}

	// Handle signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	go deathRattler(c, logger)

	// Initialize control goroutine.
	pms := plexup{running: make(chan bool)}
	go pmsController(pms.running, logger)

	// Main screen turn on.
	logger.Notice("Starting at addres: " + address)
	exec.Command("dns-sd", "-R", "plexup", "_plexup._tcp.", ".", strconv.Itoa(plexupPort), "pdl=application/plexup").Start()
	http.HandleFunc("/on", pms.on)
	http.HandleFunc("/off", pms.off)
	http.ListenAndServe(address, nil)
}
