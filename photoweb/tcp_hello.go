package main

import (
	"net/http"
	"log"
	"io"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "hello!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
}

func main() {
	http.HandleFunc("/tcp_hello", helloHandler)
	ok := http.ListenAndServe(":8000", nil)
	if ok != nil {
		log.Fatal("listenandserver: ", ok.Error())

	}
}