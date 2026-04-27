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
	quotaMaxPerPath     int
	quotaMaxPerSeverity int
)

func init() {
	quotaCmd := &cobra.Command{
		Use:   "quota",
		Short: "Check lease counts against configured quotas",
		Long:  "Evaluate active leases against per-path and per-severity limits and report violations.",
		RunE:  runQuota,
	}
	quotaCmd.Flags().IntVar(&quotaMaxPerPath, "max-per-path", 20, "Maximum leases allowed per path prefix")
	quotaCmd.Flags().IntVar(&quotaMaxPerSeverity, "max-per-severity", 50, "Maximum leases allowed per severity level")
	rootCmd.AddCommand(quotaCmd)
}

func runQuota(cmd *cobra.Command, args []string) error {
	leases := staticQuotaLeases()
	opts := filter.QuotaOptions{
		MaxPerPath:     quotaMaxPerPath,
		MaxPerSeverity: quotaMaxPerSeverity,
	}
	violations := filter.ApplyQuota(leases, opts)
	filter.PrintQuota(violations, os.Stdout)
	if len(violations) > 0 {
		return fmt.Errorf("%d quota violation(s) detected", len(violations))
	}
	return nil
}

// staticQuotaLeases returns example leases for demonstration purposes.
func staticQuotaLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "lease/secret/app/db", Path: "secret/app/db", Severity: "critical", ExpiresAt: now.Add(30 * time.Minute)},
		{LeaseID: "lease/secret/app/api", Path: "secret/app/api", Severity: "warn", ExpiresAt: now.Add(2 * time.Hour)},
		{LeaseID: "lease/secret/app/cache", Path: "secret/app/cache", Severity: "ok", ExpiresAt: now.Add(6 * time.Hour)},
	}
}
