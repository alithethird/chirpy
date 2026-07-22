package main

import (
	"net/http"
)

type apiHandler struct{}

func (apiHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

func main() {
	mux := http.NewServeMux()
	srv := http.Server{Addr: ":8080", Handler: mux}
	mux.Handle("/", http.FileServer(http.Dir(".")))
	for {
		srv.ListenAndServe()
	}
}
