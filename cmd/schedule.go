package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/vault"
)

func init() {
	var lookahead string
	var minSeverity string

	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Show upcoming alert schedule for expiring leases",
		RunE: func(cmd *cobra.Command, args []string) error {
			dur, err := time.ParseDuration(lookahead)
			if err != nil {
				return fmt.Errorf("invalid lookahead duration: %w", err)
			}

			// Static demo leases; replace with real fetcher in integration.
			leases := []vault.SecretLease{}

			entries := filter.BuildSchedule(leases, dur)
			if minSeverity != "" {
				entries = filter.FilterScheduleByMinSeverity(entries, minSeverity)
			}

			if len(entries) == 0 {
				fmt.Println("No leases scheduled for alerting in the given window.")
				return nil
			}

			for _, e := range entries {
				fmt.Println(e.String())
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&lookahead, "lookahead", "1h", "Duration window to check for expiring leases (e.g. 30m, 2h)")
	cmd.Flags().StringVar(&minSeverity, "min-severity", "", "Minimum severity to include (ok, warn, critical)")
	rootCmd.AddCommand(cmd)
}
