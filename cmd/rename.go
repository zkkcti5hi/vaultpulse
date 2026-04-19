package cmd

import (
	"fmt"
	"os"

	"github.com/nicholasgasior/vaultpulse/internal/filter"
	"github.com/nicholasgasior/vaultpulse/internal/vault"
	"github.com/spf13/cobra"
)

var renamePairs []string

var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Apply display aliases to lease paths",
	Long: `Rename lease paths using old=new pairs for cleaner output.

Example:
  vaultpulse rename --alias secret/db=database --alias secret/api=api-key`,
	RunE: runRename,
}

func init() {
	renameCmd.Flags().StringArrayVar(&renamePairs, "alias", nil, "path alias in old=new format (repeatable)")
	rootCmd.AddCommand(renameCmd)
}

func runRename(cmd *cobra.Command, args []string) error {
	aliases := filter.ParseRenameFlag(renamePairs)
	if len(aliases) == 0 {
		fmt.Fprintln(os.Stderr, "no aliases provided; use --alias old=new")
		return nil
	}

	// Placeholder: in a real invocation this would come from the fetcher.
	leases := []vault.SecretLease{}

	renamed := filter.Rename(leases, aliases)
	for _, l := range renamed {
		fmt.Printf("%s\t%s\t%s\n", l.LeaseID, l.Path, l.Severity)
	}
	return nil
}
