package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const githubWebhookSecretEnv = "GITHUB_WEBHOOK_SECRET"
const (
	defaultGitHubWebhookQueueSize   = 128
	defaultGitHubWebhookWorkerCount = 4
)

type githubWebhookDelivery struct {
	deliveryID string
	eventType  string
	payload    []byte
	receivedAt time.Time
}

type githubWebhookQueue struct {
	deliveries chan githubWebhookDelivery
}

var (
	githubWebhookQueueMu     sync.RWMutex
	activeGitHubWebhookQueue *githubWebhookQueue
)

func registerWebhookGitRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/webhook/git", webhookGitHandler)
}

func webhookGitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	secret := os.Getenv(githubWebhookSecretEnv)
	if secret == "" {
		log.Printf("missing %s environment variable", githubWebhookSecretEnv)
		http.Error(w, "webhook secret is not configured", http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	signature := r.Header.Get("X-Hub-Signature-256")
	if !verifyGitHubWebhookSignature(secret, body, signature) {
		log.Printf("invalid GitHub webhook signature")
		http.Error(w, "invalid webhook signature", http.StatusUnauthorized)
		return
	}

	delivery := githubWebhookDelivery{
		deliveryID: r.Header.Get("X-GitHub-Delivery"),
		eventType:  r.Header.Get("X-GitHub-Event"),
		payload:    append([]byte(nil), body...),
		receivedAt: time.Now().UTC(),
	}
	if !getGitHubWebhookQueue().enqueue(delivery) {
		log.Printf("GitHub webhook queue is full")
		http.Error(w, "webhook queue is full", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte("accepted"))
}

func verifyGitHubWebhookSignature(secret string, body []byte, signature string) bool {
	if signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

func newGitHubWebhookQueue(size int, workers int) *githubWebhookQueue {
	queue := &githubWebhookQueue{
		deliveries: make(chan githubWebhookDelivery, size),
	}

	for range workers {
		go queue.worker()
	}

	return queue
}

func (q *githubWebhookQueue) enqueue(delivery githubWebhookDelivery) bool {
	select {
	case q.deliveries <- delivery:
		return true
	default:
		return false
	}
}

func (q *githubWebhookQueue) worker() {
	for delivery := range q.deliveries {
		processGitHubWebhookDelivery(delivery)
	}
}

func (q *githubWebhookQueue) close() {
	close(q.deliveries)
}

func getGitHubWebhookQueue() *githubWebhookQueue {
	githubWebhookQueueMu.RLock()
	queue := activeGitHubWebhookQueue
	githubWebhookQueueMu.RUnlock()
	if queue != nil {
		return queue
	}

	githubWebhookQueueMu.Lock()
	defer githubWebhookQueueMu.Unlock()
	if activeGitHubWebhookQueue == nil {
		activeGitHubWebhookQueue = newGitHubWebhookQueue(defaultGitHubWebhookQueueSize, defaultGitHubWebhookWorkerCount)
	}

	return activeGitHubWebhookQueue
}

func setGitHubWebhookQueueForTest(queue *githubWebhookQueue) func() {
	githubWebhookQueueMu.Lock()
	previous := activeGitHubWebhookQueue
	activeGitHubWebhookQueue = queue
	githubWebhookQueueMu.Unlock()

	return func() {
		if queue != nil {
			queue.close()
		}
		githubWebhookQueueMu.Lock()
		activeGitHubWebhookQueue = previous
		githubWebhookQueueMu.Unlock()
	}
}

func processGitHubWebhookDelivery(delivery githubWebhookDelivery) {
	log.Printf(
		"processing GitHub webhook delivery=%q event=%q received_at=%s payload=%s",
		delivery.deliveryID,
		delivery.eventType,
		delivery.receivedAt.Format(time.RFC3339Nano),
		string(delivery.payload),
	)
}
