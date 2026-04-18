package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/vault"
)

var (
	exportFormat   string
	exportSeverity string
	exportPath     string
)

func init() {
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export secret leases to CSV, JSON, or text",
		RunE:  runExport,
	}
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "text", "output format: text, csv, json")
	exportCmd.Flags().StringVar(&exportSeverity, "severity", "", "minimum severity filter (ok, warning, critical)")
	exportCmd.Flags().StringVar(&exportPath, "path", "", "path prefix filter")
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, _ []string) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}
	client, err := vault.NewClientV2(cfg)
	if err != nil {
		return err
	}
	fetcher := vault.NewSecretFetcher(client)
	leases, err := fetcher.Fetch(cmd.Context())
	if err != nil {
		return err
	}
	leases = filter.Apply(leases, filter.Options{
		MinSeverity: exportSeverity,
		PathPrefix:  exportPath,
	})
	return filter.Export(os.Stdout, leases, filter.Format(exportFormat))
}
