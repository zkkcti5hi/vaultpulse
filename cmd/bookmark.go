package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpulse/internal/filter"
)

// bookmarkStore is a package-level store shared across bookmark subcommands.
var bookmarkStore = filter.NewBookmarkStore()

var bookmarkCmd = &cobra.Command{
	Use:   "bookmark",
	Short: "Manage named lease snapshots",
}

var bookmarkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved bookmarks",
	RunE: func(cmd *cobra.Command, args []string) error {
		names := bookmarkStore.List()
		if len(names) == 0 {
			fmt.Fprintln(os.Stdout, "no bookmarks saved")
			return nil
		}
		for _, n := range names {
			b, _ := bookmarkStore.Get(n)
			fmt.Fprintf(os.Stdout, "%-20s  saved=%s  leases=%d\n",
				b.Name, b.SavedAt.Format("2006-01-02T15:04:05"), len(b.Leases))
		}
		return nil
	},
}

var bookmarkDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a saved bookmark",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if _, err := bookmarkStore.Get(name); err != nil {
			return fmt.Errorf("bookmark %q not found", name)
		}
		bookmarkStore.Delete(name)
		fmt.Fprintf(os.Stdout, "bookmark %q deleted\n", name)
		return nil
	},
}

func init() {
	bookmarkCmd.AddCommand(bookmarkListCmd)
	bookmarkCmd.AddCommand(bookmarkDeleteCmd)
	rootCmd.AddCommand(bookmarkCmd)
}
