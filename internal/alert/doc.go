// Package alert provides notification mechanisms for VaultPulse lease expiration events.
//
// It includes:
//   - Notifier: writes human-readable alerts to stdout based on severity thresholds.
//   - WebhookSender: posts structured JSON payloads to a configurable HTTP endpoint
//     whenever leases require attention.
//
// Both components operate on []vault.Lease values produced by the monitor pipeline.
package alert
