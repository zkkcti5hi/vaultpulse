package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

var (
	velocityWindow   time.Duration
	velocityMinCount int
)

func init() {
	velocityCmd := &cobra.Command{
		Use:   "velocity",
		Short: "Show lease expiry velocity grouped by path prefix",
		Long: `Computes how many leases expire within a rolling window and reports
the rate per hour for each path prefix. Useful for identifying bursts of
expiring secrets that may require immediate attention.`,
		RunE: runVelocity,
	}

	velocityCmd.Flags().DurationVar(&velocityWindow, "window", 24*time.Hour,
		"Time window to measure expiry velocity (e.g. 6h, 24h)")
	velocityCmd.Flags().IntVar(&velocityMinCount, "min", 1,
		"Minimum number of leases required to include a prefix")

	rootCmd.AddCommand(velocityCmd)
}

func runVelocity(cmd *cobra.Command, args []string) error {
	leases := staticVelocityLeases()

	opts := filter.VelocityOptions{
		WindowSize: velocityWindow,
		MinLeases:  velocityMinCount,
	}

	results := filter.Velocity(leases, opts)
	if len(results) == 0 {
		fmt.Println("no leases expiring within the specified window")
		return nil
	}

	filter.PrintVelocity(results, cmd.OutOrStdout())
	return nil
}

// staticVelocityLeases returns demo leases for standalone use.
func staticVelocityLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "lease-1", Path: "secret/app/db", TTL: int((2 * time.Hour).Seconds()), Severity: "critical", CreatedAt: now},
		{LeaseID: "lease-2", Path: "secret/app/api", TTL: int((5 * time.Hour).Seconds()), Severity: "warn", CreatedAt: now},
		{LeaseID: "lease-3", Path: "secret/infra/tls", TTL: int((12 * time.Hour).Seconds()), Severity: "warn", CreatedAt: now},
		{LeaseID: "lease-4", Path: "secret/infra/ssh", TTL: int((20 * time.Hour).Seconds()), Severity: "ok", CreatedAt: now},
		{LeaseID: "lease-5", Path: "secret/app/cache", TTL: int((3 * time.Hour).Seconds()), Severity: "critical", CreatedAt: now},
	}
}
