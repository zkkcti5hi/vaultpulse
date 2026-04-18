// Package filter provides utilities for filtering, sorting, searching,
// deduplicating, paginating, grouping, summarizing, and exporting
// Vault secret lease data.
//
// Export writes a slice of SecretLease values to any io.Writer in
// CSV, JSON, or plain-text format, suitable for piping into other
// tools or saving to disk.
package filter
