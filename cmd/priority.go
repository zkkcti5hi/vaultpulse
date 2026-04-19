package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/report"
	"github.com/your-org/vaultpulse/internal/vault"
)

var priorityRules []string
var priorityFormat string

func init() {
	pCmd := &cobra.Command{
		Use:   "priority",
		Short: "Reorder leases by severity and custom priority rules",
		Long: `Apply priority rules to reorder leases. Rules are specified as
path_prefix:tag:boost triples, e.g. "prod/:vip:100".`,
		RunE: runPriority,
	}
	pCmd.Flags().StringSliceVar(&priorityRules, "rule", nil, "priority rules as prefix:tag:boost")
	pCmd.Flags().StringVar(&priorityFormat, "format", "table", "output format: table or json")
	rootCmd.AddCommand(pCmd)
}

func runPriority(cmd *cobra.Command, args []string) error {
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

	annotated := vault.Annotate(leases, cfg.Thresholds)

	rules, err := filter.ParsePriorityRules(priorityRules)
	if err != nil {
		return fmt.Errorf("parse rules: %w", err)
	}

	ordered := filter.ApplyPriority(annotated, rules)

	r := report.NewReporter(os.Stdout, priorityFormat)
	return r.Render(ordered)
}
