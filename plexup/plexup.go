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

var pms_cmd exec.Cmd
var logger syslog.Writer

func PlexOn(w http.ResponseWriter, req *http.Request) {
	logger.Notice("Turning Plex Media Server on.")
	err := pms_cmd.Start()
	io.WriteString(w, "on handler\n")
}

func PlexOff(w http.ResponseWriter, req *http.Request) {
	logger.Notice("Turning Plex Media Server off.")
	io.WriteString(w, "off handler\n")
}

func main() {
	logger, err := syslog.New(syslog.LOG_USER|syslog.LOG_NOTICE, logging_tag)
	if err != nil {
		log.Fatalln(err)
	}
	logger.Notice("Starting at addres: " + address)
	pms_cmd = exec.Cmd{
		Path: "/Applications/Plex Media Server.app/Contents/MacOS/Plex Media Server",
		Dir:  "/",
	}
	http.HandleFunc("/on", PlexOn)
	http.HandleFunc("/off", PlexOff)
	http.ListenAndServe(address, nil)
}
