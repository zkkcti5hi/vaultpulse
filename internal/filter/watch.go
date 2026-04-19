package filter

import (
	"context"
	"time"

	"github.com/vaultpulse/internal/vault"
)

// WatchConfig holds configuration for the lease watcher.
type WatchConfig struct {
	Interval  time.Duration
	Severity  string
	PathPrefix string
}

// WatchEvent is emitted each poll cycle.
type WatchEvent struct {
	Leases    []vault.SecretLease
	Changes   DiffResult
	Timestamp time.Time
}

// Watcher polls a fetcher and emits WatchEvents on changes.
type Watcher struct {
	cfg     WatchConfig
	fetch   func() ([]vault.SecretLease, error)
	events  chan WatchEvent
	prev    []vault.SecretLease
}

// NewWatcher creates a Watcher that calls fetch on each interval.
func NewWatcher(cfg WatchConfig, fetch func() ([]vault.SecretLease, error)) *Watcher {
	if cfg.Interval <= 0 {
		cfg.Interval = 30 * time.Second
	}
	return &Watcher{
		cfg:    cfg,
		fetch:  fetch,
		events: make(chan WatchEvent, 8),
	}
}

// Events returns the read-only event channel.
func (w *Watcher) Events() <-chan WatchEvent {
	return w.events
}

// Run starts the polling loop; blocks until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			leases, err := w.fetch()
			if err != nil {
				continue
			}
			filtered := Apply(leases, w.cfg.Severity, w.cfg.PathPrefix)
			changes := Diff(w.prev, filtered)
			w.prev = filtered
			w.events <- WatchEvent{
				Leases:    filtered,
				Changes:   changes,
				Timestamp: time.Now(),
			}
		}
	}
}
