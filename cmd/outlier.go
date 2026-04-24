package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/vault"
)

var (
	outlierMultiplier float64
)

func init() {
	outlierCmd := &cobra.Command{
		Use:   "outlier",
		Short: "Detect leases with abnormally low TTL relative to peers",
		Long: `Computes the mean and standard deviation of all lease TTLs.
Leases whose TTL falls more than --multiplier standard deviations below
the mean are reported as outliers.`,
		RunE: runOutlier,
	}
	outlierCmd.Flags().Float64Var(&outlierMultiplier, "multiplier", 2.0,
		"Standard deviation multiplier for outlier threshold")
	rootCmd.AddCommand(outlierCmd)
}

func runOutlier(cmd *cobra.Command, args []string) error {
	leases := staticOutlierLeases()
	if len(leases) == 0 {
		fmt.Fprintln(os.Stdout, "No leases to analyse.")
		return nil
	}
	results := filter.DetectOutliers(leases, outlierMultiplier)
	filter.PrintOutliers(results, os.Stdout)
	return nil
}

// staticOutlierLeases returns a small set of demo leases for the outlier
// command. In production this would be replaced by a real Vault fetch.
func staticOutlierLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "lease/db", Path: "secret/db", TTL: 3600 * time.Second, ExpiresAt: now.Add(3600 * time.Second), Severity: "ok"},
		{LeaseID: "lease/api", Path: "secret/api", TTL: 3500 * time.Second, ExpiresAt: now.Add(3500 * time.Second), Severity: "ok"},
		{LeaseID: "lease/cache", Path: "secret/cache", TTL: 3700 * time.Second, ExpiresAt: now.Add(3700 * time.Second), Severity: "ok"},
		{LeaseID: "lease/mq", Path: "secret/mq", TTL: 3600 * time.Second, ExpiresAt: now.Add(3600 * time.Second), Severity: "ok"},
		{LeaseID: "lease/short", Path: "secret/short", TTL: 30 * time.Second, ExpiresAt: now.Add(30 * time.Second), Severity: "critical"},
	}
}
