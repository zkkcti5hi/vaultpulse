package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/vault"
)

func init() {
	var window string
	var minSeverity string

	cmd := &cobra.Command{
		Use:   "forecast",
		Short: "Show leases predicted to expire within a time window",
		Long: `Forecast scans all known leases and lists those expected to expire
within the specified look-ahead window, ordered soonest-first.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runForecast(window, minSeverity)
		},
	}

	cmd.Flags().StringVar(&window, "window", "72h", "Look-ahead duration (e.g. 24h, 48h, 72h)")
	cmd.Flags().StringVar(&minSeverity, "min-severity", "warn", "Minimum severity to include (ok, warn, critical)")

	rootCmd.AddCommand(cmd)
}

func runForecast(windowStr, minSeverity string) error {
	d, err := time.ParseDuration(windowStr)
	if err != nil {
		return fmt.Errorf("invalid --window %q: %w", windowStr, err)
	}

	leases := staticForecastLeases()

	opts := filter.ForecastOptions{
		Window:      d,
		MinSeverity: minSeverity,
	}

	entries := filter.Forecast(leases, opts)
	filter.PrintForecast(entries, os.Stdout)
	return nil
}

// staticForecastLeases returns demo leases for the forecast command.
// In production this would be replaced by a live Vault fetch.
func staticForecastLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "lease/db", Path: "secret/prod/db", Severity: "critical", LeaseDuration: int(now.Add(30 * time.Minute).Unix() - now.Unix())},
		{LeaseID: "lease/api", Path: "secret/prod/api", Severity: "warn", LeaseDuration: int(now.Add(6 * time.Hour).Unix() - now.Unix())},
		{LeaseID: "lease/cache", Path: "secret/prod/cache", Severity: "ok", LeaseDuration: int(now.Add(96 * time.Hour).Unix() - now.Unix())},
	}
}
