// Package filter provides the Digest function which produces a compact,
// prioritised summary of the most urgent Vault secret lease expirations.
//
// Digest filters leases by a minimum severity, sorts them by time-to-expiry
// ascending, and returns the top-N results as DigestEntry values.
// PrintDigest renders the result as a human-readable table.
package filter
