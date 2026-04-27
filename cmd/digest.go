package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

func init() {
	var topN int
	var minSev string

	digestCmd := &cobra.Command{
		Use:   "digest",
		Short: "Print a compact digest of the most urgent expiring leases",
		Long: `digest filters leases by minimum severity, sorts by time-to-expiry,
and prints the top-N results in a concise table.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDigest(topN, minSev)
		},
	}

	digestCmd.Flags().IntVar(&topN, "top", 10, "number of leases to display")
	digestCmd.Flags().StringVar(&minSev, "min-severity", "warn", "minimum severity (ok|warn|critical)")

	rootCmd.AddCommand(digestCmd)
}

func runDigest(topN int, minSev string) error {
	leases := staticDigestLeases()

	opts := filter.DefaultDigestOptions()
	opts.TopN = topN
	opts.MinSev = minSev
	opts.Writer = os.Stdout

	entries := filter.Digest(leases, opts)
	filter.PrintDigest(entries, os.Stdout)
	return nil
}

// staticDigestLeases returns demo leases for CLI use.
// In production this would be replaced by a real Vault fetcher.
func staticDigestLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "lease-001", Path: "secret/db/prod", Severity: "critical", ExpiresAt: now.Add(8 * time.Minute)},
		{LeaseID: "lease-002", Path: "secret/api/key", Severity: "warn", ExpiresAt: now.Add(45 * time.Minute)},
		{LeaseID: "lease-003", Path: "secret/tls/cert", Severity: "ok", ExpiresAt: now.Add(72 * time.Hour)},
		{LeaseID: "lease-004", Path: "secret/db/staging", Severity: "warn", ExpiresAt: now.Add(2 * time.Hour)},
	}
}

var _ = fmt.Sprintf // suppress unused import
