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
	clusterStrategy string
)

func init() {
	clusterCmd := &cobra.Command{
		Use:   "cluster",
		Short: "Group leases into clusters by prefix, severity, or tag",
		Long: `Cluster displays leases grouped by a chosen strategy.

Strategies:
  prefix    - group by top-level path segment (default)
  severity  - group by severity label
  tag       - group by tag combination`,
		RunE: runCluster,
	}

	clusterCmd.Flags().StringVarP(&clusterStrategy, "strategy", "s", "prefix",
		"clustering strategy: prefix | severity | tag")

	rootCmd.AddCommand(clusterCmd)
}

func runCluster(cmd *cobra.Command, args []string) error {
	leases := staticClusterLeases()
	if len(leases) == 0 {
		fmt.Fprintln(os.Stderr, "no leases available")
		return nil
	}

	valid := map[string]bool{"prefix": true, "severity": true, "tag": true}
	if !valid[clusterStrategy] {
		return fmt.Errorf("unknown strategy %q: choose prefix, severity, or tag", clusterStrategy)
	}

	clusters := filter.ClusterBy(leases, clusterStrategy)
	filter.PrintClusters(clusters, os.Stdout)
	return nil
}

// staticClusterLeases returns demo leases for the cluster command.
func staticClusterLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "id1", Path: "secret/app/db", Severity: "critical", Tags: []string{"prod"}, ExpiresAt: now.Add(10 * time.Minute)},
		{LeaseID: "id2", Path: "secret/app/api", Severity: "warn", Tags: []string{"prod"}, ExpiresAt: now.Add(2 * time.Hour)},
		{LeaseID: "id3", Path: "auth/token", Severity: "ok", Tags: []string{"dev"}, ExpiresAt: now.Add(24 * time.Hour)},
		{LeaseID: "id4", Path: "secret/infra/tls", Severity: "critical", Tags: []string{"prod", "infra"}, ExpiresAt: now.Add(5 * time.Minute)},
	}
}
