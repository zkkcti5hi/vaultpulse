package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/vaultpulse/internal/config"
	"github.com/user/vaultpulse/internal/report"
	"github.com/user/vaultpulse/internal/vault"
)

var reportFormat string

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Print a snapshot of current secret lease expirations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		client, err := vault.NewClientV2(cfg)
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		fetcher := vault.NewSecretFetcher(client, cfg)
		leases, err := fetcher.Fetch(cmd.Context())
		if err != nil {
			return fmt.Errorf("fetch leases: %w", err)
		}

		annotated := vault.Annotate(leases, cfg)

		fmt := report.Format(reportFormat)
		r := report.NewReporter(os.Stdout, fmt)
		return r.Write(annotated)
	},
}

func init() {
	reportCmd.Flags().StringVarP(&reportFormat, "format", "f", "table", "Output format: table or json")
	rootCmd.AddCommand(reportCmd)
}
