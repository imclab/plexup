package main

import (
	"io"
	"log"
	"log/syslog"
	"net/http"
	"os/exec"
)

const logging_tag = "plexup"
const address = ":25010"

type plexup struct {
	cmd    exec.Cmd
	logger *syslog.Writer
}

func (p *plexup) on(w http.ResponseWriter, req *http.Request) {
	p.logger.Notice("Turning Plex Media Server on.")
	p.cmd.Start()
	// http://127.0.0.1:32400/library/sections/all/refresh
	io.WriteString(w, "on handler\n")
}

func (p *plexup) off(w http.ResponseWriter, req *http.Request) {
	p.logger.Notice("Turning Plex Media Server off.")
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
	pms.cmd = exec.Cmd{
		Path: "/Applications/Plex Media Server.app/Contents/MacOS/Plex Media Server",
		Dir:  "/",
	}
	http.HandleFunc("/on", pms.on)
	http.HandleFunc("/off", pms.off)
	http.ListenAndServe(address, nil)
}
