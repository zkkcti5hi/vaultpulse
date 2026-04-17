package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/you/vaultpulse/internal/audit"
	"github.com/you/vaultpulse/internal/config"
	"github.com/you/vaultpulse/internal/vault"
)

var auditOutput string

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Print a JSON audit log of current lease expiration states",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		client, err := vault.NewClientV2(cfg)
		if err != nil {
			return err
		}
		fetcher := vault.NewSecretFetcher(client)
		leases, err := fetcher.Fetch(cmd.Context())
		if err != nil {
			return err
		}
		annotated := vault.Annotate(leases, cfg.Alerts)

		out := os.Stdout
		if auditOutput != "" && auditOutput != "-" {
			f, err := os.OpenFile(auditOutput, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				return err
			}
			defer f.Close()
			out = f
		}

		logger := audit.NewLogger(out)
		return logger.Log(annotated)
	},
}

func init() {
	auditCmd.Flags().StringVarP(&auditOutput, "output", "o", "-", "File path for audit log output (- for stdout)")
	rootCmd.AddCommand(auditCmd)
}
