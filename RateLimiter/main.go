package main

import (
	"log"
	"net/http"
)

const DEFAULT_HOST = "localhost:8989"

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /unlimited", unlimitedHandler)
	mux.HandleFunc("GET /limited", limitedHandler)

	log.Fatal(http.ListenAndServe(DEFAULT_HOST, mux))
}

func unlimitedHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("<a>Unlimited! Let's Go!</a>"))
}

func limitedHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("<a>Limited, don't over use me!</a>"))
}
