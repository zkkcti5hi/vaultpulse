package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "vaultpulse",
	Short: "Monitor HashiCorp Vault secret lease expirations",
	Long: `vaultpulse watches Vault secret leases and alerts you
before they expire, giving you time to renew or rotate secrets.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.yaml", "Path to config file")
}
