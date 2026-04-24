// Package filter provides lease filtering, transformation, and analysis
// utilities for the vaultpulse CLI.
//
// The outlier sub-feature identifies secret leases whose remaining TTL deviates
// significantly from the statistical mean TTL across all monitored leases.
// Outliers are surfaced so operators can investigate leases that may have been
// misconfigured or are expiring unexpectedly soon relative to their peers.
package filter
