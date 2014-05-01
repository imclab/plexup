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
	"time"
)

// loggingTag is used as a tag for syslog messages.
const loggingTag = "plexup"

// plexupPort sets the default port plexup will launch HTTP interface on.
const plexupPort = 25010

// plexup is a control structure used to control PMS. It's a simple way of
// sharing the control channel between HTTP handlers.
type plexup struct {
	running chan bool
}

// pmsController is a goroutine that starts or stops PMS.
func pmsController(c chan bool, logger *syslog.Writer) {
	// finished is a non-buffered channel used for actual PMS control.
	// Instead of launcing PMS and tracking the pid, I launch a goroutine that
	// launches PMS and waits for it to finish. Once it does, it sends a dummy
	// value (struct{}) down this channel.
	finished := make(chan struct{}, 1)
	finished <- struct{}{}
	for {
		desiredState := <-c
		switch desiredState {
		case true:
			logger.Notice("Turning Plex Media Server on.")
			select {
			// This branch runs if finished is empty - that is PMS is still running.
			default:
				logger.Notice("...although I think Plex Media Server is already running.")
			case <-finished:
				exec.Command("/usr/bin/pmset", "relative", "wake", "1").Run()
				exec.Command("/usr/bin/osascript", "-e", "try", "-e", "mount volume \"smb://yacoob@boxoob/worek\"", "-e", "end try").Run()
				cmd := exec.Command("/usr/bin/caffeinate", "/Applications/Plex Media Server.app/Contents/MacOS/Plex Media Server")
				go func() {
					cmd.Run()
					finished <- struct{}{}
				}()
				// Wait a moment to allow PMS to initialize.
				time.Sleep(2 * time.Second)
				// Request a refresh of the library.
				http.Get("http://127.0.0.1:32400/library/sections/all/refresh")
				logger.Notice("Plex Media Server started.")
			}
		case false:
			logger.Notice("Turning Plex Media Server off.")
			gpid, _ := syscall.Getpgid(os.Getpid())
			// TODO: avoid calling pkill
			exec.Command("/usr/bin/pkill", "-g", strconv.Itoa(gpid), "Plex Media Server").Run()
		}
	}
}

func (pms *plexup) on(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "on handler\n")
	// TODO: actually output something from pmsController to the client
	pms.running <- true
}

func (pms *plexup) off(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "off handler\n")
	pms.running <- false
}

func (pms *plexup) quit(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "bye!\n")
	quit()
}

// quit kills whole process group.
func quit() {
	syscall.Kill(0, syscall.SIGTERM)
	os.Exit(1)
}

// deathRatller should run on plexup receiving SIGTERM - it'll try to clean up
// all launched processes.
func deathRattler(c chan os.Signal, logger *syslog.Writer) {
	<-c
	logger.Notice("SIGTERM received - killing all processes in my process group.")
	quit()
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
	exec.Command("/usr/bin/dns-sd", "-R", "plexup", "_plexup._tcp.", ".", strconv.Itoa(plexupPort), "pdl=application/plexup").Start()
	http.HandleFunc("/on", pms.on)
	http.HandleFunc("/off", pms.off)
	http.HandleFunc("/quit", pms.quit)
	http.ListenAndServe(address, nil)
}
