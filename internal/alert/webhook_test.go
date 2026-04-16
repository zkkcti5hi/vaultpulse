package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeWebhookLease(severity string) vault.Lease {
	return vault.Lease{
		LeaseID:   "secret/data/test#abc",
		Path:      "secret/data/test",
		ExpiresAt: time.Now().Add(2 * time.Hour),
		TTL:       2 * time.Hour,
		Severity:  severity,
	}
}

func TestWebhookSender_SendsPayload(t *testing.T) {
	var received WebhookPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sender := NewWebhookSender(ts.URL)
	leases := []vault.Lease{makeWebhookLease("critical"), makeWebhookLease("warning")}
	if err := sender.Send(leases); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received.Leases) != 2 {
		t.Errorf("expected 2 leases, got %d", len(received.Leases))
	}
	if received.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestWebhookSender_EmptyLeasesNoRequest(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer ts.Close()

	sender := NewWebhookSender(ts.URL)
	if err := sender.Send(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected no HTTP request for empty leases")
	}
}

func TestWebhookSender_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	sender := NewWebhookSender(ts.URL)
	leases := []vault.Lease{makeWebhookLease("critical")}
	if err := sender.Send(leases); err == nil {
		t.Error("expected error for non-OK status")
	}
}
