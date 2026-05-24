package main

import "net/http"

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	registerHealthRoutes(mux)
	registerWebhookRoutes(mux)
	registerWebhookGitRoutes(mux)
	return mux
}
