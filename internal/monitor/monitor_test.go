package monitor_test

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/vaultpulse/internal/alert"
	"github.com/your-org/vaultpulse/internal/monitor"
	"github.com/your-org/vaultpulse/internal/vault"
)

// stubClient satisfies the interface used by Runner via duck-typing in tests.
type stubClient struct {
	leases []vault.Lease
	calls  int
}

func (s *stubClient) ListLeases(_ context.Context) ([]vault.Lease, error) {
	s.calls++
	return s.leases, nil
}

func makeNotifier() *alert.Notifier {
	return alert.NewNotifier(alert.Options{
		CriticalThreshold: 24 * time.Hour,
		WarningThreshold:  72 * time.Hour,
	})
}

func TestNewRunner_DefaultInterval(t *testing.T) {
	n := makeNotifier()
	r := monitor.NewRunner(nil, n, 0)
	if r == nil {
		t.Fatal("expected non-nil runner")
	}
}

func TestRunner_CancelStopsLoop(t *testing.T) {
	n := makeNotifier()
	// Use a very short interval so the loop ticks quickly.
	r := monitor.NewRunner(nil, n, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// poll will fail because client is nil, but the loop should still exit cleanly.
	err := r.Run(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}
