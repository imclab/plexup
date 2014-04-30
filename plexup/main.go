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

// Idea from gobry: one channel-triggered control gouroutine

type plexup struct {
	logger   *syslog.Writer
	finished chan struct{}
}

func (pms *plexup) on(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "on handler\n")
	select {
	default:
		io.WriteString(w, "I think pms is already running.\n")
		return
	case <-pms.finished:
		pms.logger.Notice("Turning Plex Media Server on.")
	}
	cmd := exec.Command("/usr/bin/caffeinate", "/Applications/Plex Media Server.app/Contents/MacOS/Plex Media Server")
	go func() {
		cmd.Run()
		pms.finished <- struct{}{}
	}()
	// TODO: hit http://127.0.0.1:32400/library/sections/all/refresh
}

func (pms *plexup) off(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "off handler\n")
	pms.logger.Notice("Turning Plex Media Server off.")
	gpid, _ := syscall.Getpgid(os.Getpid())
	// TODO: avoid calling pkill
	exec.Command("pkill", "-g", strconv.Itoa(gpid), "Plex Media Server").Run()
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

	// Initialize control structure.
	var pms = new(plexup)
	pms.finished = make(chan struct{}, 1)
	pms.finished <- struct{}{}
	if pms.logger, err = syslog.New(syslog.LOG_USER|syslog.LOG_NOTICE, loggingTag); err != nil {
		log.Fatalln(err)
	}

	// Main screen turn on.
	pms.logger.Notice("Starting at addres: " + address)

	// Handle signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	go deathRattler(c, pms.logger)

	exec.Command("dns-sd", "-R", "plexup", "_plexup._tcp.", ".", strconv.Itoa(plexupPort), "pdl=application/plexup").Start()
	http.HandleFunc("/on", pms.on)
	http.HandleFunc("/off", pms.off)
	http.ListenAndServe(address, nil)
}
