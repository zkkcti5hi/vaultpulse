package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpulse/internal/config"
	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/report"
	"github.com/yourusername/vaultpulse/internal/vault"
)

var (
	searchQuery         string
	searchCaseSensitive bool
	searchFormat        string
)

func init() {
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search leases by ID or path substring",
		RunE:  runSearch,
	}
	searchCmd.Flags().StringVarP(&searchQuery, "query", "q", "", "substring to search for (required)")
	searchCmd.Flags().BoolVar(&searchCaseSensitive, "case-sensitive", false, "enable case-sensitive matching")
	searchCmd.Flags().StringVarP(&searchFormat, "format", "f", "table", "output format: table or json")
	_ = searchCmd.MarkFlagRequired("query")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
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

	leases = vault.Annotate(leases, cfg.Alerts)
	leases = filter.Search(leases, filter.SearchOptions{
		Query:         searchQuery,
		CaseSensitive: searchCaseSensitive,
	})

	r := report.NewReporter(os.Stdout, searchFormat)
	if err := r.Render(leases); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	return nil
}
