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
	var field string

	cmd := &cobra.Command{
		Use:   "correlation",
		Short: "Group leases by shared attributes to surface correlated expirations",
		Long: `Correlate groups leases that share a common field (path-prefix, severity, tag).
Groups with more than one member are printed so operators can spot clusters
of related expirations that may require coordinated renewal.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCorrelation(field)
		},
	}

	cmd.Flags().StringVar(&field, "field", "path-prefix",
		"Field to correlate by: path-prefix | severity | tag")

	rootCmd.AddCommand(cmd)
}

func runCorrelation(field string) error {
	leases := staticCorrelationLeases()
	r := filter.Correlate(leases, field)
	if len(r.Groups) == 0 {
		fmt.Fprintln(os.Stdout, "No correlated groups found.")
		return nil
	}
	filter.PrintCorrelation(r, os.Stdout)
	return nil
}

// staticCorrelationLeases returns demo leases for the correlation command.
// In production this would be replaced by a real Vault fetch.
func staticCorrelationLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "lease-1", Path: "secret/app/db", Severity: "critical", ExpiresAt: now.Add(10 * time.Minute)},
		{LeaseID: "lease-2", Path: "secret/app/cache", Severity: "critical", ExpiresAt: now.Add(15 * time.Minute)},
		{LeaseID: "lease-3", Path: "secret/infra/tls", Severity: "warn", ExpiresAt: now.Add(2 * time.Hour)},
		{LeaseID: "lease-4", Path: "secret/infra/ssh", Severity: "warn", ExpiresAt: now.Add(3 * time.Hour)},
		{LeaseID: "lease-5", Path: "secret/other/svc", Severity: "ok", ExpiresAt: now.Add(24 * time.Hour)},
	}
}
