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
	retentionMaxAge string
	retentionMaxTTL string
)

func init() {
	retentionCmd := &cobra.Command{
		Use:   "retention",
		Short: "Flag leases that violate retention policies",
		Long: `Inspect leases and report any that exceed configured maximum age
or TTL thresholds, helping enforce secret hygiene policies.`,
		RunE: runRetention,
	}

	retentionCmd.Flags().StringVar(&retentionMaxAge, "max-age", "720h",
		"Maximum allowed lease age since first seen (e.g. 24h, 30d)")
	retentionCmd.Flags().StringVar(&retentionMaxTTL, "max-ttl", "2160h",
		"Maximum allowed lease TTL (e.g. 24h, 90d)")

	rootCmd.AddCommand(retentionCmd)
}

func runRetention(cmd *cobra.Command, args []string) error {
	maxAge, err := time.ParseDuration(retentionMaxAge)
	if err != nil {
		return fmt.Errorf("invalid --max-age %q: %w", retentionMaxAge, err)
	}
	maxTTL, err := time.ParseDuration(retentionMaxTTL)
	if err != nil {
		return fmt.Errorf("invalid --max-ttl %q: %w", retentionMaxTTL, err)
	}

	leases := staticRetentionLeases()

	policy := filter.RetentionPolicy{
		MaxAge: maxAge,
		MaxTTL: maxTTL,
	}

	results := filter.ApplyRetention(leases, policy)
	filter.PrintRetention(results, os.Stdout)
	return nil
}

// staticRetentionLeases returns example leases for demonstration.
// In production this would be replaced by a real Vault fetch.
func staticRetentionLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "lease/a", Path: "secret/db/prod", TTL: 86400, SeenAt: now.Add(-45 * 24 * time.Hour), Severity: "warn"},
		{LeaseID: "lease/b", Path: "secret/api/key", TTL: 3600, SeenAt: now.Add(-2 * time.Hour), Severity: "ok"},
		{LeaseID: "lease/c", Path: "secret/infra/tls", TTL: 100 * 86400, SeenAt: now.Add(-1 * time.Hour), Severity: "critical"},
	}
}
