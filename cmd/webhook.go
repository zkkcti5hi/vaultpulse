package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpulse/internal/alert"
	"github.com/yourusername/vaultpulse/internal/config"
	"github.com/yourusername/vaultpulse/internal/vault"
)

var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Send current lease alerts to a configured webhook URL",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgFile, _ := cmd.Flags().GetString("config")
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		if cfg.WebhookURL == "" {
			return fmt.Errorf("webhook_url is not set in config")
		}

		client, err := vault.NewClientV2(cfg)
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}
		fetcher := vault.NewSecretFetcher(client)
		leases, err := fetcher.Fetch(cmd.Context())
		if err != nil {
			return fmt.Errorf("fetch leases: %w", err)
		}
		annotated := vault.Annotate(leases, cfg.Thresholds.Warning, cfg.Thresholds.Critical)

		sender := alert.NewWebhookSender(cfg.WebhookURL)
		if err := sender.Send(annotated); err != nil {
			log.Printf("webhook error: %v", err)
			return err
		}
		fmt.Printf("Webhook delivered: %d lease(s) sent to %s\n", len(annotated), cfg.WebhookURL)
		return nil
	},
}

func init() {
	webhookCmd.Flags().String("config", "", "path to config file")
	RootCmd.AddCommand(webhookCmd)
}
