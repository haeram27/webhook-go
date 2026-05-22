package main

import "net/http"

func registerHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", healthzHandler)
}

func healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
