// Package monitor implements the core polling loop for VaultPulse.
//
// Runner periodically queries HashiCorp Vault for active secret leases,
// annotates each lease with an expiry severity level (using vault.Annotate),
// and forwards the results to the configured alert.Notifier.
//
// Typical usage:
//
//	client, _ := vault.NewClientV2(cfg)
//	notifier := alert.NewNotifier(alert.Options{...})
//	runner := monitor.NewRunner(client, notifier, 60*time.Second)
//	runner.Run(ctx)
package monitor
