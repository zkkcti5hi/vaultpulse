package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vaultpulse/internal/filter"
	"github.com/vaultpulse/internal/report"
	"github.com/vaultpulse/internal/vault"
)

var (
	filterSeverity   string
	filterPathPrefix string
	filterFormat     string
)

func init() {
	filterCmd := &cobra.Command{
		Use:   "filter",
		Short: "Filter and display leases by severity or path prefix",
		RunE:  runFilter,
	}
	filterCmd.Flags().StringVar(&filterSeverity, "severity", "", "Minimum severity: ok, warning, critical")
	filterCmd.Flags().StringVar(&filterPathPrefix, "path-prefix", "", "Filter leases by path prefix")
	filterCmd.Flags().StringVar(&filterFormat, "format", "table", "Output format: table or json")
	rootCmd.AddCommand(filterCmd)
}

func runFilter(cmd *cobra.Command, args []string) error {
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

	annotated := vault.Annotate(leases, cfg.Alerts)
	filtered := filter.Apply(annotated, filter.Options{
		Severity:   filterSeverity,
		PathPrefix: filterPathPrefix,
	})

	r := report.NewReporter(os.Stdout, filterFormat)
	if err := r.Render(filtered); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	return nil
}
