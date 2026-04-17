// Package audit provides a structured audit log for lease expiration events.
package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/you/vaultpulse/internal/vault"
)

// Entry represents a single audit log record.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	LeaseID   string    `json:"lease_id"`
	Path      string    `json:"path"`
	Severity  string    `json:"severity"`
	TTL       string    `json:"ttl"`
	Message   string    `json:"message"`
}

// Logger writes audit entries to a destination.
type Logger struct {
	out io.Writer
}

// NewLogger returns a Logger writing to out. If out is nil, os.Stderr is used.
func NewLogger(out io.Writer) *Logger {
	if out == nil {
		out = os.Stderr
	}
	return &Logger{out: out}
}

// Log writes an audit entry for each lease.
func (l *Logger) Log(leases []vault.SecretLease) error {
	for _, lease := range leases {
		entry := Entry{
			Timestamp: time.Now().UTC(),
			LeaseID:   lease.LeaseID,
			Path:      lease.Path,
			Severity:  lease.Severity,
			TTL:       lease.TTLDuration().String(),
			Message:   fmt.Sprintf("lease %s expires in %s", lease.LeaseID, lease.TTLDuration()),
		}
		b, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("audit: marshal: %w", err)
		}
		if _, err := fmt.Fprintln(l.out, string(b)); err != nil {
			return fmt.Errorf("audit: write: %w", err)
		}
	}
	return nil
}
