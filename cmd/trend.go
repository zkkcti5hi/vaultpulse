package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/vault"
)

var trendCmd = &cobra.Command{
	Use:   "trend",
	Short: "Show severity trend across recent snapshots",
	RunE:  runTrend,
}

var trendSnapshotDir string

func init() {
	trendCmd.Flags().StringVar(&trendSnapshotDir, "snapshot-dir", ".vaultpulse/snapshots", "Directory containing snapshots")
	rootCmd.AddCommand(trendCmd)
}

func runTrend(cmd *cobra.Command, args []string) error {
	store, err := filter.NewSnapshotStore(trendSnapshotDir)
	if err != nil {
		return fmt.Errorf("open snapshot store: %w", err)
	}
	names, err := store.List()
	if err != nil {
		return fmt.Errorf("list snapshots: %w", err)
	}
	if len(names) == 0 {
		fmt.Fprintln(os.Stdout, "no snapshots found")
		return nil
	}
	var all [][]vault.SecretLease
	for _, name := range names {
		leases, err := store.Get(name)
		if err != nil {
			c}
		all = append(all, leases)
	}
	points := filter.Trend(all)
	fmt.Fprint(os.Stdout, filter.PrintTrend(points))
	return nil
}
