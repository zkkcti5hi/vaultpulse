package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// WebhookPayload is the JSON body sent to a webhook endpoint.
type WebhookPayload struct {
	Timestamp string        `json:"timestamp"`
	Leases    []vault.Lease `json:"leases"`
	Summary   string        `json:"summary"`
}

// WebhookSender sends lease alerts to an HTTP endpoint.
type WebhookSender struct {
	URL    string
	Client *http.Client
}

// NewWebhookSender creates a WebhookSender with a default timeout.
func NewWebhookSender(url string) *WebhookSender {
	return &WebhookSender{
		URL: url,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send posts the leases as JSON to the configured webhook URL.
func (w *WebhookSender) Send(leases []vault.Lease) error {
	if len(leases) == 0 {
		return nil
	}
	payload := WebhookPayload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Leases:    leases,
		Summary:   fmt.Sprintf("%d lease(s) require attention", len(leases)),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook marshal: %w", err)
	}
	resp, err := w.Client.Post(w.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
