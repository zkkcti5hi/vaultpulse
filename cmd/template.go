package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vaultpulse/internal/filter"
	"github.com/vaultpulse/internal/vault"
)

var (
	templateFlagStr  string
	templateFilePath string
)

func init() {
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Render lease data using a custom Go text/template",
		RunE:  runTemplate,
	}
	templateCmd.Flags().StringVarP(&templateFlagStr, "template", "t", "", "Go text/template string for rendering leases")
	templateCmd.Flags().StringVarP(&templateFilePath, "file", "f", "", "Path to a template file (overrides --template)")
	rootCmd.AddCommand(templateCmd)
}

func runTemplate(cmd *cobra.Command, args []string) error {
	leases := []vault.SecretLease{
		{LeaseID: "example/lease/1", Path: "secret/data/db", Severity: "critical", TTL: "1h"},
		{LeaseID: "example/lease/2", Path: "secret/data/api", Severity: "warn", TTL: "6h"},
	}

	tmplStr := templateFlagStr
	if templateFilePath != "" {
		data, err := os.ReadFile(templateFilePath)
		if err != nil {
			return err
		}
		tmplStr = string(data)
	}

	return filter.RenderTemplate(leases, tmplStr, os.Stdout)
}
