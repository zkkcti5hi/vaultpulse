package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/vaultpulse/internal/filter"
)

var globalPinStore = filter.NewPinStore()

func init() {
	pinCmd := &cobra.Command{
		Use:   "pin",
		Short: "Manage pinned secret leases",
	}

	pinCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all pinned leases",
		RunE: func(cmd *cobra.Command, args []string) error {
			leases := globalPinStore.List()
			if len(leases) == 0 {
				fmt.Println("No pinned leases.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "LEASE ID\tPATH\tSEVERITY\tEXPIRES AT")
			for _, l := range leases {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					l.LeaseID, l.Path, l.Severity, l.ExpiresAt.Format("2006-01-02 15:04:05"))
			}
			return w.Flush()
		},
	})

	pinCmd.AddCommand(&cobra.Command{
		Use:   "remove [lease-id]",
		Short: "Remove a pinned lease by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !globalPinStore.Unpin(args[0]) {
				return fmt.Errorf("lease %q not found in pins", args[0])
			}
			fmt.Printf("Unpinned lease %s\n", args[0])
			return nil
		},
	})

	rootCmd.AddCommand(pinCmd)
}
