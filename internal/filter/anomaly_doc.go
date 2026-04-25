// Package filter provides lease filtering, sorting, grouping, and analysis
// utilities for vaultpulse.
//
// The anomaly sub-feature detects leases that deviate from normal behaviour,
// such as those with unusually short TTLs or leases that appeared very
// recently — both of which may indicate misconfiguration or an active incident.
package filter
