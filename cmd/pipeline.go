package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/vaultpulse/internal/filter"
	"github.com/vaultpulse/internal/vault"
)

var (
	pipelineSteps    string
	pipelineMinSev   string
	pipelinePathPfx  string
)

func init() {
	pipelineCmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Run leases through a named, ordered filter pipeline",
		Long: `Execute a sequence of named filter steps against the lease list.

Available step names: filter-severity, filter-path, sort-expiry, dedupe.

Example:
  vaultpulse pipeline --steps filter-severity,sort-expiry,dedupe --min-severity warn`,
		RunE: runPipeline,
	}
	pipelineCmd.Flags().StringVar(&pipelineSteps, "steps", "filter-severity,sort-expiry,dedupe",
		"Comma-separated ordered list of pipeline step names")
	pipelineCmd.Flags().StringVar(&pipelineMinSev, "min-severity", "warn",
		"Minimum severity for the filter-severity step (ok|warn|critical)")
	pipelineCmd.Flags().StringVar(&pipelinePathPfx, "path-prefix", "",
		"Path prefix for the filter-path step")
	rootCmd.AddCommand(pipelineCmd)
}

func runPipeline(cmd *cobra.Command, _ []string) error {
	leases := staticPipelineLeases()

	p := filter.NewPipeline().WithWriter(cmd.OutOrStdout())

	for _, name := range filter.ParsePipelineSteps(pipelineSteps) {
		switch name {
		case "filter-severity":
			minSev := pipelineMinSev
			p.Add(name, func(ls []vault.SecretLease) []vault.SecretLease {
				return filter.Apply(ls, filter.Options{MinSeverity: minSev})
			})
		case "filter-path":
			pfx := pipelinePathPfx
			p.Add(name, func(ls []vault.SecretLease) []vault.SecretLease {
				return filter.Apply(ls, filter.Options{PathPrefix: pfx})
			})
		case "sort-expiry":
			p.Add(name, func(ls []vault.SecretLease) []vault.SecretLease {
				return filter.Sort(ls, filter.SortOptions{By: "expiry", Desc: false})
			})
		case "dedupe":
			p.Add(name, func(ls []vault.SecretLease) []vault.SecretLease {
				return filter.Dedupe(ls)
			})
		default:
			return fmt.Errorf("unknown pipeline step: %q", name)
		}
	}

	filter.PrintPipeline(p, cmd.OutOrStdout())
	result := p.Run(leases)
	fmt.Fprintf(cmd.OutOrStdout(), "\nFinal result: %d lease(s)\n", len(result))
	return nil
}

func staticPipelineLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "lease/db/prod", Path: "secret/db/prod", Severity: "critical", TTL: 5 * time.Minute, ExpiresAt: now.Add(5 * time.Minute)},
		{LeaseID: "lease/db/prod", Path: "secret/db/prod", Severity: "critical", TTL: 5 * time.Minute, ExpiresAt: now.Add(5 * time.Minute)},
		{LeaseID: "lease/app/token", Path: "secret/app/token", Severity: "warn", TTL: 25 * time.Minute, ExpiresAt: now.Add(25 * time.Minute)},
		{LeaseID: "lease/infra/cert", Path: "secret/infra/cert", Severity: "ok", TTL: 2 * time.Hour, ExpiresAt: now.Add(2 * time.Hour)},
	}
}

var _ = os.Stderr // ensure os imported
