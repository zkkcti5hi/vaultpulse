package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/vaultpulse/internal/filter"
)

var (
	globalSuppressStore = filter.NewSuppressStore()
	suppressDuration    string
)

var suppressCmd = &cobra.Command{
	Use:   "suppress <lease-id>",
	Short: "Suppress alerts for a lease ID for a given duration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		leaseID := args[0]
		var until time.Time
		if suppressDuration != "" {
			d, err := time.ParseDuration(suppressDuration)
			if err != nil {
				return fmt.Errorf("invalid duration %q: %w", suppressDuration, err)
			}
			until = time.Now().Add(d)
		}
		globalSuppressStore.Suppress(leaseID, until)
		if until.IsZero() {
			fmt.Fprintf(os.Stdout, "suppressed %s indefinitely\n", leaseID)
		} else {
			fmt.Fprintf(os.Stdout, "suppressed %s until %s\n", leaseID, until.Format(time.RFC3339))
		}
		return nil
	},
}

var unsuppressCmd = &cobra.Command{
	Use:   "unsuppress <lease-id>",
	Short: "Remove suppression for a lease ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		leaseID := args[0]
		if !globalSuppressStore.Unsuppress(leaseID) {
			return fmt.Errorf("lease %q not found in suppress list", leaseID)
		}
		fmt.Fprintf(os.Stdout, "unsuppressed %s\n", leaseID)
		return nil
	},
}

func init() {
	suppressCmd.Flags().StringVarP(&suppressDuration, "duration", "d", "", "suppress for duration e.g. 2h, 30m (omit for indefinite)")
	rootCmd.AddCommand(suppressCmd)
	rootCmd.AddCommand(unsuppressCmd)
}
