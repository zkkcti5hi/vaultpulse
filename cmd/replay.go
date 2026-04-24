package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

var (
	replayCount    int
	replayInterval time.Duration
)

func init() {
	replayCmd := &cobra.Command{
		Use:   "replay",
		Short: "Record and replay lease snapshots over time",
		Long: `Captures periodic lease snapshots and prints a timeline showing
how lease counts and severities changed. Useful for post-incident review.`,
		RunE: runReplay,
	}
	replayCmd.Flags().IntVarP(&replayCount, "count", "n", 5, "number of snapshots to capture")
	replayCmd.Flags().DurationVarP(&replayInterval, "interval", "i", 5*time.Second, "interval between snapshots")
	rootCmd.AddCommand(replayCmd)
}

func runReplay(cmd *cobra.Command, args []string) error {
	leases := staticReplayLeases()
	store := filter.NewReplayStore()

	fmt.Fprintf(os.Stderr, "Recording %d snapshots every %s...\n", replayCount, replayInterval)

	for i := 0; i < replayCount; i++ {
		store.Record(time.Now(), leases)
		if i < replayCount-1 {
			time.Sleep(replayInterval)
		}
	}

	filter.PrintReplay(os.Stdout, store.All())
	return nil
}

// staticReplayLeases returns demo leases for the replay command.
func staticReplayLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "lease/db/prod/1", Path: "db/prod", Severity: "critical", TTL: 300, ExpiresAt: now.Add(5 * time.Minute)},
		{LeaseID: "lease/api/token/2", Path: "api/token", Severity: "warn", TTL: 1800, ExpiresAt: now.Add(30 * time.Minute)},
		{LeaseID: "lease/infra/cert/3", Path: "infra/cert", Severity: "ok", TTL: 86400, ExpiresAt: now.Add(24 * time.Hour)},
	}
}
