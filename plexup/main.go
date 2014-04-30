package main

import (
	"io"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

const logging_tag = "plexup"
const address = ":25010"

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
	var gpid, _ = syscall.Getpgid(os.Getpid())
	var pid = strconv.Itoa(gpid)
	exec.Command("pkill", "-g", pid, "Plex Media Server").Run()
}

func main() {
	var err error

	// Initialize control structure.
	var pms = new(plexup)
	pms.finished = make(chan struct{}, 1)
	pms.finished <- struct{}{}
	if pms.logger, err = syslog.New(syslog.LOG_USER|syslog.LOG_NOTICE, logging_tag); err != nil {
		log.Fatalln(err)
	}

	// sleep proxy:
	// dns-sd -R "plexup" _plexup._tcp. . 25010 pdl=application/plexup
	// Main screen turn on.
	pms.logger.Notice("Starting at addres: " + address)
	http.HandleFunc("/on", pms.on)
	http.HandleFunc("/off", pms.off)
	http.ListenAndServe(address, nil)
}
