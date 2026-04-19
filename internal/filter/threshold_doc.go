// Package filter provides filtering, sorting, and transformation utilities
// for Vault secret leases.
//
// The threshold sub-feature allows operators to override default severity
// classification by specifying custom warn and critical TTL thresholds.
// Use ApplyThreshold to re-classify a slice of leases, and ParseThresholdFlag
// to parse user-supplied CLI flags in the form "warn=72,critical=24".
package filter
