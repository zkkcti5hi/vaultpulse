package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

var snapshotFile string

func init() {
	snapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage lease snapshots",
	}

	saveCmd := &cobra.Command{
		Use:   "save [name]",
		Short: "Save current leases as a named snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := filter.NewSnapshotStore(snapshotFile)
			// In a real run this would come from the vault fetcher; use empty for CLI demo.
			leases := []vault.SecretLease{}
			if err := store.Save(args[0], leases); err != nil {
				return fmt.Errorf("save snapshot: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Snapshot %q saved.\n", args[0])
			return nil
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all saved snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := filter.NewSnapshotStore(snapshotFile)
			snaps := store.List()
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tCREATED\tLEASES")
			for _, s := range snaps {
				fmt.Fprintf(w, "%s\t%s\t%d\n", s.Name, s.CreatedAt.Format("2006-01-02 15:04:05"), len(s.Leases))
			}
			return w.Flush()
		},
	}

	deleteCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a named snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := filter.NewSnapshotStore(snapshotFile)
			if err := store.Delete(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Snapshot %q deleted.\n", args[0])
			return nil
		},
	}

	snapshotCmd.PersistentFlags().StringVar(&snapshotFile, "store", "snapshots.json", "Path to snapshot store file")
	snapshotCmd.AddCommand(saveCmd, listCmd, deleteCmd)
	rootCmd.AddCommand(snapshotCmd)
	_ = os.MkdirAll(".", 0o755)
}
