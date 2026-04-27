package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/vaultpulse/internal/filter"
	"github.com/vaultpulse/internal/vault"
)

func init() {
	var windowSize string
	var numWindows int

	cmd := &cobra.Command{
		Use:   "window",
		Short: "Show lease expirations grouped into rolling time windows",
		Long: `Partitions leases into sequential time buckets starting from now.
Each bucket covers a fixed duration (e.g. 1h) and shows how many
leases expire within that window.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWindow(windowSize, numWindows)
		},
	}

	cmd.Flags().StringVar(&windowSize, "window", "1h", "size of each time bucket (e.g. 30m, 1h, 6h)")
	cmd.Flags().IntVar(&numWindows, "num", 6, "number of time buckets to display")

	rootCmd.AddCommand(cmd)
}

func runWindow(windowSize string, numWindows int) error {
	dur, err := time.ParseDuration(windowSize)
	if err != nil {
		return fmt.Errorf("invalid --window duration %q: %w", windowSize, err)
	}
	if numWindows < 1 {
		return fmt.Errorf("--num must be at least 1")
	}

	leases := staticWindowLeases()

	opts := filter.DefaultWindowOptions()
	opts.WindowSize = dur
	opts.NumWindows = numWindows
	opts.Out = os.Stdout

	buckets := filter.RollingWindow(leases, opts)
	filter.PrintWindows(buckets, opts)
	return nil
}

// staticWindowLeases returns demo leases for the window command.
// Replace with a real Vault client fetch in production.
func staticWindowLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "lease/db/prod", Path: "db/prod", ExpiresAt: now.Add(20 * time.Minute)},
		{LeaseID: "lease/db/staging", Path: "db/staging", ExpiresAt: now.Add(45 * time.Minute)},
		{LeaseID: "lease/api/key", Path: "api/key", ExpiresAt: now.Add(90 * time.Minute)},
		{LeaseID: "lease/tls/cert", Path: "tls/cert", ExpiresAt: now.Add(3 * time.Hour)},
		{LeaseID: "lease/ssh/host", Path: "ssh/host", ExpiresAt: now.Add(5 * time.Hour)},
	}
}
