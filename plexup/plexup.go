package main

import (
	"io"
	"log"
	"log/syslog"
	"net/http"
)

const logging_tag = "plexup"
const address = ":25010"

var logger syslog.Writer

func PlexOn(w http.ResponseWriter, req *http.Request) {
	logger.Notice("Turning Plex Media Server on.")
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
	http.HandleFunc("/on", PlexOn)
	http.HandleFunc("/off", PlexOff)
	http.ListenAndServe(address, nil)
}
