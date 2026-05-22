package main

import (
	"io"
	"log"
	"net/http"
)

func registerWebhookRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/webhook", webhookHandler)
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("received webhook payload: %s", string(body))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("received"))
}
