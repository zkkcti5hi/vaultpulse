package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/report"
	"github.com/yourusername/vaultpulse/internal/vault"
)

var (
	tagList   []string
	tagFormat string
)

func init() {
	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Filter leases by path tags",
		Long:  "Filter secret leases whose path contains one or more specified tags (case-insensitive substring match).",
		RunE:  runTag,
	}
	tagCmd.Flags().StringSliceVarP(&tagList, "tags", "t", nil, "Comma-separated list of tags to filter by (e.g. db,prod)")
	tagCmd.Flags().StringVarP(&tagFormat, "format", "f", "table", "Output format: table or json")
	_ = tagCmd.MarkFlagRequired("tags")
	rootCmd.AddCommand(tagCmd)
}

func runTag(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("config: %w", err)
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
	filtered := filter.FilterByTags(annotated, tagList)
	r := report.NewReporter(report.Options{Format: tagFormat})
	fmt.Fprintf(cmd.OutOrStdout(), "Leases matching tags [%s]:\n", strings.Join(tagList, ", "))
	return r.Render(cmd.OutOrStdout(), filtered)
}
