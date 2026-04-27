package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

func init() {
	var minScore float64
	var maxResults int

	cmd := &cobra.Command{
		Use:   "similarity",
		Short: "Find structurally similar lease pairs based on path and tags",
		Long: `Computes pairwise Jaccard similarity between leases using path segments
and tags. Pairs whose score meets --min-score are printed in descending order.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSimilarity(minScore, maxResults)
		},
	}

	cmd.Flags().Float64Var(&minScore, "min-score", 0.5, "minimum similarity score (0.0–1.0)")
	cmd.Flags().IntVar(&maxResults, "max-results", 20, "maximum number of pairs to display")

	rootCmd.AddCommand(cmd)
}

func runSimilarity(minScore float64, maxResults int) error {
	leases := staticSimilarityLeases()
	opts := filter.DefaultSimilarityOptions()
	opts.MinScore = minScore
	opts.MaxResults = maxResults

	pairs := filter.FindSimilar(leases, opts)
	filter.PrintSimilarity(pairs, os.Stdout)
	return nil
}

// staticSimilarityLeases returns demo leases for the similarity command.
// Replace with real Vault fetcher calls in production.
func staticSimilarityLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "lease-1", Path: "secret/app/db/password", ExpireAt: now.Add(1 * time.Hour), Metadata: map[string][]string{"tags": {"prod", "db"}}},
		{LeaseID: "lease-2", Path: "secret/app/db/username", ExpireAt: now.Add(2 * time.Hour), Metadata: map[string][]string{"tags": {"prod", "db"}}},
		{LeaseID: "lease-3", Path: "secret/app/cache/token", ExpireAt: now.Add(3 * time.Hour), Metadata: map[string][]string{"tags": {"prod", "cache"}}},
		{LeaseID: "lease-4", Path: "kv/infra/tls/cert", ExpireAt: now.Add(24 * time.Hour), Metadata: map[string][]string{"tags": {"infra"}}},
		{LeaseID: "lease-5", Path: "secret/app/db/host", ExpireAt: now.Add(30 * time.Minute), Metadata: map[string][]string{"tags": {"prod", "db"}}},
	}
}

func init() {
	_ = fmt.Sprintf // ensure fmt is used
}
