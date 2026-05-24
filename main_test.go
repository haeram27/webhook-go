package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

func TestWebhookGitPost(t *testing.T) {
	t.Setenv(githubWebhookSecretEnv, "test-secret")
	restoreQueue := setGitHubWebhookQueueForTest(newGitHubWebhookQueue(1, 0))
	defer restoreQueue()

	payload := `{"event":"push"}`
	req := httptest.NewRequest(http.MethodPost, "/webhook/git", strings.NewReader(payload))
	req.Header.Set("X-Hub-Signature-256", githubTestSignature("test-secret", payload))
	rr := httptest.NewRecorder()

	newMux().ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rr.Code)
	}
}

func TestWebhookGitRejectsInvalidSignature(t *testing.T) {
	t.Setenv(githubWebhookSecretEnv, "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/webhook/git", strings.NewReader(`{"event":"push"}`))
	req.Header.Set("X-Hub-Signature-256", "sha256=invalid")
	rr := httptest.NewRecorder()

	newMux().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestWebhookGitRejectsMissingSignature(t *testing.T) {
	t.Setenv(githubWebhookSecretEnv, "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/webhook/git", strings.NewReader(`{"event":"push"}`))
	rr := httptest.NewRecorder()

	newMux().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestWebhookGitRequiresSecret(t *testing.T) {
	t.Setenv(githubWebhookSecretEnv, "")

	req := httptest.NewRequest(http.MethodPost, "/webhook/git", strings.NewReader(`{"event":"push"}`))
	req.Header.Set("X-Hub-Signature-256", "sha256=invalid")
	rr := httptest.NewRecorder()

	newMux().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestWebhookGitReturnsServiceUnavailableWhenQueueIsFull(t *testing.T) {
	t.Setenv(githubWebhookSecretEnv, "test-secret")
	restoreQueue := setGitHubWebhookQueueForTest(newGitHubWebhookQueue(1, 0))
	defer restoreQueue()

	payload := `{"event":"push"}`
	firstReq := httptest.NewRequest(http.MethodPost, "/webhook/git", strings.NewReader(payload))
	firstReq.Header.Set("X-Hub-Signature-256", githubTestSignature("test-secret", payload))
	firstRes := httptest.NewRecorder()

	newMux().ServeHTTP(firstRes, firstReq)

	if firstRes.Code != http.StatusAccepted {
		t.Fatalf("expected first status %d, got %d", http.StatusAccepted, firstRes.Code)
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/webhook/git", strings.NewReader(payload))
	secondReq.Header.Set("X-Hub-Signature-256", githubTestSignature("test-secret", payload))
	secondRes := httptest.NewRecorder()

	newMux().ServeHTTP(secondRes, secondReq)

	if secondRes.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, secondRes.Code)
	}
}

func TestVerifyGitHubWebhookSignatureMatchesGitHubExample(t *testing.T) {
	secret := "It's a Secret to Everybody"
	payload := "Hello, World!"
	signature := "sha256=757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17"

	if !verifyGitHubWebhookSignature(secret, []byte(payload), signature) {
		t.Fatal("expected signature to match GitHub documentation example")
	}
}

func githubTestSignature(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
