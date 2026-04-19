package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

var scoreCmd = &cobra.Command{
	Use:   "score",
	Short: "Rank leases by risk score",
	Long:  "Compute and display a risk score for each lease, ordered highest first.",
	RunE:  runScore,
}

func init() {
	rootCmd.AddCommand(scoreCmd)
}

func runScore(cmd *cobra.Command, args []string) error {
	leases := []vault.SecretLease{} // placeholder: wire real fetcher
	scored := filter.Score(leases)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SCORE\tSEVERITY\tPATH\tLEASE ID")
	for _, s := range scored {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n",
			s.Score,
			s.Lease.Severity,
			s.Lease.Path,
			s.Lease.LeaseID,
		)
	}
	return w.Flush()
}
