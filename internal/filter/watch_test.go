package filter_test

import (
	"context"
	"testing"
	"time"

	"github.com/vaultpulse/internal/filter"
	"github.com/vaultpulse/internal/vault"
)

func makeWatchLease(id, path, severity string, ttl time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:  id,
		Path:     path,
		Severity: severity,
		TTL:      ttl,
	}
}

func TestWatcher_EmitsEvent(t *testing.T) {
	leases := []vault.SecretLease{
		makeWatchLease("id1", "secret/a", "critical", time.Minute),
	}
	fetch := func() ([]vault.SecretLease, error) { return leases, nil }

	w := filter.NewWatcher(filter.WatchConfig{Interval: 20 * time.Millisecond}, fetch)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go w.Run(ctx) //nolint:errcheck

	select {
	case ev := <-w.Events():
		if len(ev.Leases) != 1 {
			t.Fatalf("expected 1 lease, got %d", len(ev.Leases))
		}
		if ev.Timestamp.IsZero() {
			t.Error("expected non-zero timestamp")
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for watch event")
	}
}

func TestWatcher_DefaultInterval(t *testing.T) {
	w := filter.NewWatcher(filter.WatchConfig{}, func() ([]vault.SecretLease, error) {
		return nil, nil
	})
	// Just ensure construction doesn't panic; interval defaulted internally.
	if w == nil {
		t.Fatal("expected non-nil watcher")
	}
}

func TestWatcher_CancelStops(t *testing.T) {
	fetch := func() ([]vault.SecretLease, error) { return nil, nil }
	w := filter.NewWatcher(filter.WatchConfig{Interval: 10 * time.Millisecond}, fetch)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- w.Run(ctx) }()
	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("watcher did not stop after cancel")
	}
}
