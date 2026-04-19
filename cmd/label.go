package cmd

import (
	"fmt"
	"strings"

	"github.com/nicholasgasior/vaultpulse/internal/filter"
	"github.com/spf13/cobra"
)

var labelStore = filter.NewLabelStore()

var labelCmd = &cobra.Command{
	Use:   "label",
	Short: "Manage labels on leases",
}

var labelAddCmd = &cobra.Command{
	Use:   "add <leaseID> <label>",
	Short: "Add a label to a lease",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		labelStore.Add(args[0], args[1])
		fmt.Printf("Label %q added to lease %q\n", args[1], args[0])
		return nil
	},
}

var labelRemoveCmd = &cobra.Command{
	Use:   "remove <leaseID> <label>",
	Short: "Remove a label from a lease",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		labelStore.Remove(args[0], args[1])
		fmt.Printf("Label %q removed from lease %q\n", args[1], args[0])
		return nil
	},
}

var labelListCmd = &cobra.Command{
	Use:   "list <leaseID>",
	Short: "List labels for a lease",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		labels := labelStore.Get(args[0])
		if len(labels) == 0 {
			fmt.Printf("No labels for lease %q\n", args[0])
			return nil
		}
		fmt.Printf("Labels for %q: %s\n", args[0], strings.Join(labels, ", "))
		return nil
	},
}

func init() {
	labelCmd.AddCommand(labelAddCmd, labelRemoveCmd, labelListCmd)
	rootCmd.AddCommand(labelCmd)
}
