// Package audit provides structured JSON audit logging for VaultPulse lease
// expiration events. Each monitored lease that crosses a severity threshold is
// recorded with a timestamp, lease ID, path, severity level, remaining TTL,
// and a human-readable message. Audit entries are written one JSON object per
// line to any io.Writer, making them easy to ship to log aggregation systems.
package audit
