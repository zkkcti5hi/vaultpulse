package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/report"
	"github.com/your-org/vaultpulse/internal/vault"
)

var thresholdFlag string
var thresholdFormat string

func init() {
	thresholdCmd := &cobra.Command{
		Use:   "threshold",
		Short: "Re-classify leases using custom TTL thresholds",
		Long: `Re-annotates lease severities based on custom warn and critical
TTL thresholds. Accepts a flag in the form "warn=72,critical=24" (hours).`,
		RunE: runThreshold,
	}
	thresholdCmd.Flags().StringVar(&thresholdFlag, "thresholds", "", `Custom thresholds, e.g. "warn=72,critical=24"`)
	thresholdCmd.Flags().StringVar(&thresholdFormat, "format", "table", "Output format: table or json")
	rootCmd.AddCommand(thresholdCmd)
}

func runThreshold(cmd *cobra.Command, args []string) error {
	cfg, err := filter.ParseThresholdFlag(thresholdFlag)
	if err != nil {
		return fmt.Errorf("threshold flag: %w", err)
	}

	// Use static demo leases; in production these come from the Vault fetcher.
	leases := []vault.SecretLease{}
	result := filter.ApplyThreshold(leases, cfg)

	r := report.NewReporter(os.Stdout, thresholdFormat)
	if err := r.Render(result); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	return nil
}
