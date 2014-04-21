package main

import (
	"io"
	"log"
	"log/syslog"
	"net/http"
)

const logging_tag = "plexup"
const address = ":25010"

func OHai(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "ohai\n")
}

func main() {
	logger, err := syslog.New(syslog.LOG_USER|syslog.LOG_NOTICE, logging_tag)
	if err != nil {
		log.Fatalln(err)
	}
	logger.Notice("Starting at addres: " + address)
	http.HandleFunc("/", OHai)
	http.ListenAndServe(address, nil)
}
