package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebhookPost(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(`{"event":"ping"}`))
	rr := httptest.NewRecorder()

	newMux().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestWebhookMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	rr := httptest.NewRecorder()

	newMux().ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
