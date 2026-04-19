package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/filter/snapshot"
)

var diffCmd = &cobra.Command{
	Use:   "diff <snapshot-a> <snapshot-b>",
	Short: "Diff two saved snapshots and show added/removed/changed leases",
	Args:  cobra.ExactArgs(2),
	RunE:  runDiff,
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	store, err := filter.NewSnapshotStore(snapshotDir())
	if err != nil {
		return fmt.Errorf("open snapshot store: %w", err)
	}

	before, err := store.Get(args[0])
	if err != nil {
		return fmt.Errorf("load snapshot %q: %w", args[0], err)
	}
	after, err := store.Get(args[1])
	if err != nil {
		return fmt.Errorf("load snapshot %q: %w", args[1], err)
	}

	d := filter.Diff(before, after)
	if d.IsEmpty() {
		fmt.Fprintln(os.Stdout, "no differences found")
		return nil
	}
	fmt.Fprintf(os.Stdout, "Summary: %s\n\n", d.Summary())
	filter.PrintDiff(os.Stdout, d)
	return nil
}

func snapshotDir() string {
	if d := os.Getenv("VAULTPULSE_SNAPSHOT_DIR"); d != "" {
		return d
	}
	return ".vaultpulse/snapshots"
}
