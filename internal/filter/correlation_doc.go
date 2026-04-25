// Package filter provides lease filtering, grouping, and analysis utilities
// for the vaultpulse CLI.
//
// Correlation identifies leases that share a common attribute — such as a
// path prefix, severity level, or tag — and groups them so operators can
// quickly spot clusters of related expirations that may indicate a systemic
// issue or coordinated renewal requirement.
package filter
