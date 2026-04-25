package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/vaultpulse/internal/filter"
	"github.com/user/vaultpulse/internal/vault"
)

func init() {
	var shortTTL time.Duration
	var recentWindow time.Duration

	cmd := &cobra.Command{
		Use:   "anomaly",
		Short: "Detect anomalous leases based on TTL and appearance time",
		Long: `Scans leases for anomalous behaviour such as unusually short TTLs
or leases that appeared very recently, which may indicate misconfiguration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnomaly(shortTTL, recentWindow)
		},
	}

	cmd.Flags().DurationVar(&shortTTL, "short-ttl", 5*time.Minute,
		"Flag leases with TTL below this threshold")
	cmd.Flags().DurationVar(&recentWindow, "recent-window", 10*time.Minute,
		"Flag leases first seen within this window")

	rootCmd.AddCommand(cmd)
}

func runAnomaly(shortTTL, recentWindow time.Duration) error {
	leases := staticAnomalyLeases()

	opts := filter.AnomalyOptions{
		ShortTTLThreshold:  shortTTL,
		RecentlySeenWindow: recentWindow,
	}

	results := filter.DetectAnomalies(leases, opts)
	filter.PrintAnomalies(results, os.Stdout)

	if len(results) > 0 {
		fmt.Fprintf(os.Stderr, "%d anomal(ies) detected.\n", len(results))
	}
	return nil
}

// staticAnomalyLeases returns demo leases for CLI usage.
func staticAnomalyLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "lease-1", Path: "secret/db", Severity: "critical", ExpiresAt: now.Add(3 * time.Minute)},
		{LeaseID: "lease-2", Path: "secret/api", Severity: "warn", ExpiresAt: now.Add(2 * time.Hour), SeenAt: now.Add(-2 * time.Minute)},
		{LeaseID: "lease-3", Path: "secret/cache", Severity: "ok", ExpiresAt: now.Add(24 * time.Hour)},
	}
}
