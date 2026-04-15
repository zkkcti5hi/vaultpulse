package monitor

import (
	"context"
	"log"
	"time"

	"github.com/your-org/vaultpulse/internal/alert"
	"github.com/your-org/vaultpulse/internal/vault"
)

// Runner periodically polls Vault for lease data and triggers alerts.
type Runner struct {
	client   *vault.ClientV2
	notifier *alert.Notifier
	interval time.Duration
}

// NewRunner creates a new monitor Runner.
func NewRunner(client *vault.ClientV2, notifier *alert.Notifier, interval time.Duration) *Runner {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	return &Runner{
		client:   client,
		notifier: notifier,
		interval: interval,
	}
}

// Run starts the polling loop and blocks until ctx is cancelled.
func (r *Runner) Run(ctx context.Context) error {
	log.Printf("[monitor] starting poll loop (interval=%s)", r.interval)
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	// Run once immediately before waiting for the first tick.
	if err := r.poll(ctx); err != nil {
		log.Printf("[monitor] poll error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("[monitor] shutting down")
			return ctx.Err()
		case <-ticker.C:
			if err := r.poll(ctx); err != nil {
				log.Printf("[monitor] poll error: %v", err)
			}
		}
	}
}

// poll fetches leases, annotates them, and sends alerts.
func (r *Runner) poll(ctx context.Context) error {
	leases, err := r.client.ListLeases(ctx)
	if err != nil {
		return err
	}

	annotated := vault.Annotate(leases)
	log.Printf("[monitor] fetched %d leases, %d annotated", len(leases), len(annotated))

	r.notifier.Notify(annotated)
	return nil
}
