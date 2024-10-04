package main

import (
	"log"
	"net/http"
)

func main() {
	const address = "localhost:8080"
	mux := http.NewServeMux()

	mux.HandleFunc("/_stats", statsHandler)

	log.Printf("listening on %s...", address)
	err := http.ListenAndServe(address, mux)
	if err != nil {
		panic(err)
	}
}

func statsHandler(response http.ResponseWriter, request *http.Request) {
	_, _ = response.Write([]byte("12,21474836481934,1073741824,5497558138880,4398046511104,104857600,6291456\n"))
}
