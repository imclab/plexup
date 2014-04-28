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

type plexup struct {
	logger *syslog.Writer
}

func (p *plexup) on(w http.ResponseWriter, req *http.Request) {
	p.logger.Notice("Turning Plex Media Server on.")
	exec.Command("/usr/bin/caffeinate", "/Applications/Plex Media Server.app/Contents/MacOS/Plex Media Server").Start()
	// http://127.0.0.1:32400/library/sections/all/refresh
	io.WriteString(w, "on handler\n")
}

func (p *plexup) off(w http.ResponseWriter, req *http.Request) {
	p.logger.Notice("Turning Plex Media Server off.")
	var gpid, _ = syscall.Getpgid(os.Getpid())
	var pid = strconv.Itoa(gpid)
	exec.Command("pkill", "-g", pid, "Plex Media Server").Run()
	io.WriteString(w, "off handler\n")
}

func main() {
	var pms = new(plexup)
	logger, err := syslog.New(syslog.LOG_USER|syslog.LOG_NOTICE, logging_tag)
	if err != nil {
		log.Fatalln(err)
	}
	pms.logger = logger
	pms.logger.Notice("Starting at addres: " + address)
	http.HandleFunc("/on", pms.on)
	http.HandleFunc("/off", pms.off)
	http.ListenAndServe(address, nil)
}
