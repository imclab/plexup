package main

import (
	"errors"
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

type plexup struct {
	logger  *syslog.Writer
	watcher chan error
}

func (p *plexup) on(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "on handler\n")
	select {
	default:
		io.WriteString(w, "I think pms is already running.\n")
		return
	case _ = <-p.watcher:
		p.logger.Notice("Turning Plex Media Server on.")
	}
	cmd := exec.Command("/usr/bin/caffeinate", "/Applications/Plex Media Server.app/Contents/MacOS/Plex Media Server")
	go func() {
		p.watcher <- cmd.Run()
	}()
	// http://127.0.0.1:32400/library/sections/all/refresh
}

func (p *plexup) off(w http.ResponseWriter, req *http.Request) {
	p.logger.Notice("Turning Plex Media Server off.")
	var gpid, _ = syscall.Getpgid(os.Getpid())
	var pid = strconv.Itoa(gpid)
	exec.Command("pkill", "-g", pid, "Plex Media Server").Run()
	io.WriteString(w, "off handler\n")
}

func main() {
	var err error
	var pms = new(plexup)

	// Initialize control structure.
	pms.watcher = make(chan error, 1)
	pms.watcher <- errors.New("Cleared for launch.")
	pms.logger, err = syslog.New(syslog.LOG_USER|syslog.LOG_NOTICE, logging_tag)
	if err != nil {
		log.Fatalln(err)
	}

	// Main screen turn on.
	pms.logger.Notice("Starting at addres: " + address)
	http.HandleFunc("/on", pms.on)
	http.HandleFunc("/off", pms.off)
	http.ListenAndServe(address, nil)
}
