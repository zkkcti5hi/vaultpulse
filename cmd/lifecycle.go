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
	lifecycleWarnWindow string
	lifecycleStage      string
)

func init() {
	lifecycleCmd := &cobra.Command{
		Use:   "lifecycle",
		Short: "Classify and display secret leases by lifecycle stage",
		Long: `Classify leases as active, expiring, expired, or renewing.

Examples:
  vaultpulse lifecycle --warn-window 24h
  vaultpulse lifecycle --stage expiring`,
		RunE: runLifecycle,
	}
	lifecycleCmd.Flags().StringVar(&lifecycleWarnWindow, "warn-window", "24h", "duration before expiry to classify as 'expiring'")
	lifecycleCmd.Flags().StringVar(&lifecycleStage, "stage", "", "filter output to a specific stage (active|expiring|expired|renewing)")
	rootCmd.AddCommand(lifecycleCmd)
}

func runLifecycle(cmd *cobra.Command, args []string) error {
	warnWindow, err := time.ParseDuration(lifecycleWarnWindow)
	if err != nil {
		return fmt.Errorf("invalid --warn-window %q: %w", lifecycleWarnWindow, err)
	}

	leases := staticLeases()
	entries := filter.ClassifyLifecycle(leases, warnWindow)

	if lifecycleStage != "" {
		stage := filter.LifecycleStage(lifecycleStage)
		entries = filter.FilterByStage(entries, stage)
	}

	filter.PrintLifecycle(entries, os.Stdout)
	return nil
}

// staticLeases returns a placeholder slice; in production this would come from
// the configured Vault client via the secret fetcher.
func staticLeases() []vault.SecretLease {
	return []vault.SecretLease{}
}
