package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/report"
	"github.com/yourusername/vaultpulse/internal/vault"
)

var dedupeFormat string

var dedupeCmd = &cobra.Command{
	Use:   "dedupe",
	Short: "Remove duplicate leases by lease ID and report unique leases",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		client, err := vault.NewClientV2(cfg)
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}
		fetcher := vault.NewSecretFetcher(client)
		leases, err := fetcher.Fetch(cmd.Context())
		if err != nil {
			return fmt.Errorf("fetch leases: %w", err)
		}
		leases = filter.Dedupe(leases)
		r := report.NewReporter(os.Stdout, dedupeFormat)
		return r.Render(leases)
	},
}

func init() {
	dedupeCmd.Flags().StringVarP(&dedupeFormat, "format", "f", "table", "Output format: table or json")
	rootCmd.AddCommand(dedupeCmd)
}
