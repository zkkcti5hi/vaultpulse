package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vaultpulse/internal/filter"
)

var (
	noteStore   = filter.NewNoteStore()
	noteLeaseID string
	noteText    string
)

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Manage notes attached to secret leases",
}

var noteSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Attach a note to a lease",
	RunE: func(cmd *cobra.Command, args []string) error {
		if noteLeaseID == "" || noteText == "" {
			return fmt.Errorf("--lease and --note are required")
		}
		noteStore.Set(noteLeaseID, noteText)
		fmt.Fprintf(os.Stdout, "Note set for lease %s\n", noteLeaseID)
		return nil
	},
}

var noteGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve the note for a lease",
	RunE: func(cmd *cobra.Command, args []string) error {
		if noteLeaseID == "" {
			return fmt.Errorf("--lease is required")
		}
		n, ok := noteStore.Get(noteLeaseID)
		if !ok {
			fmt.Fprintf(os.Stdout, "No note found for lease %s\n", noteLeaseID)
			return nil
		}
		fmt.Fprintln(os.Stdout, n)
		return nil
	},
}

var noteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all leases with notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		ids := noteStore.List()
		if len(ids) == 0 {
			fmt.Fprintln(os.Stdout, "No notes stored.")
			return nil
		}
		for _, id := range ids {
			n, _ := noteStore.Get(id)
			fmt.Fprintf(os.Stdout, "%s: %s\n", id, n)
		}
		return nil
	},
}

func init() {
	noteSetCmd.Flags().StringVar(&noteLeaseID, "lease", "", "Lease ID to annotate")
	noteSetCmd.Flags().StringVar(&noteText, "note", "", "Note text")
	noteGetCmd.Flags().StringVar(&noteLeaseID, "lease", "", "Lease ID to look up")
	noteCmd.AddCommand(noteSetCmd, noteGetCmd, noteListCmd)
	rootCmd.AddCommand(noteCmd)
}
