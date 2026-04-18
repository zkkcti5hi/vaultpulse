package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpulse/internal/filter"
)

var globalHistory = filter.NewHistory(20)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Display previously recorded lease snapshots",
	RunE:  runHistory,
}

func init() {
	historyCmd.Flags().IntP("limit", "n", 0, "Show only the last N snapshots (0 = all)")
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, _ []string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	snapshots := globalHistory.All()
	if len(snapshots) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No snapshots recorded yet.")
		return nil
	}
	if limit > 0 && limit < len(snapshots) {
		snapshots = snapshots[len(snapshots)-limit:]
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "#\tCAPTURED AT\tLEASE COUNT")
	for i, s := range snapshots {
		fmt.Fprintf(w, "%d\t%s\t%d\n",
			i+1,
			s.CapturedAt.Format(time.RFC3339),
			len(s.Leases),
		)
	}
	return w.Flush()
}
