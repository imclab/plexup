package main

import (
	"io"
	"net/http"
)

const address = ":25010"

func OHai(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "ohai\n")
}

func main() {
	http.HandleFunc("/", OHai)
	http.ListenAndServe(address, nil)
}
