package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

var topologyMinSeverity string

func init() {
	topologyCmd := &cobra.Command{
		Use:   "topology",
		Short: "Display secret paths as a hierarchical tree",
		Long:  `Builds and prints a tree view of all monitored secret paths, grouped by path segment.`,
		RunE:  runTopology,
	}
	topologyCmd.Flags().StringVar(&topologyMinSeverity, "min-severity", "ok", "Minimum severity to include (ok, warn, critical)")
	rootCmd.AddCommand(topologyCmd)
}

func runTopology(cmd *cobra.Command, args []string) error {
	leases := staticTopologyLeases()
	filtered := filter.Apply(leases, filter.Options{MinSeverity: topologyMinSeverity})
	if len(filtered) == 0 {
		fmt.Fprintln(os.Stdout, "No leases match the given filters.")
		return nil
	}
	root := filter.BuildTopology(filtered)
	filter.PrintTopology(root, os.Stdout)
	return nil
}

// staticTopologyLeases returns example leases for demo/test purposes.
func staticTopologyLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "a", Path: "secret/app/db", TTL: 30 * time.Minute, ExpiresAt: now.Add(30 * time.Minute), Severity: "critical"},
		{LeaseID: "b", Path: "secret/app/cache", TTL: 2 * time.Hour, ExpiresAt: now.Add(2 * time.Hour), Severity: "warn"},
		{LeaseID: "c", Path: "kv/infra/tls", TTL: 24 * time.Hour, ExpiresAt: now.Add(24 * time.Hour), Severity: "ok"},
		{LeaseID: "d", Path: "kv/infra/ssh", TTL: 12 * time.Hour, ExpiresAt: now.Add(12 * time.Hour), Severity: "ok"},
	}
}
