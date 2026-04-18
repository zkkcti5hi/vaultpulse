// Package filter provides lease filtering utilities for vaultpulse.
package filter

import (
	"strings"

	"github.com/vaultpulse/internal/vault"
)

// Options holds filter criteria for leases.
type Options struct {
	// Severity filters leases to only those at or above this level.
	// Accepted values: "ok", "warning", "critical". Empty means no filter.
	Severity string

	// PathPrefix filters leases whose path starts with the given prefix.
	PathPrefix string
}

// Apply returns a subset of leases matching all non-empty criteria in opts.
func Apply(leases []vault.SecretLease, opts Options) []vault.SecretLease {
	result := make([]vault.SecretLease, 0, len(leases))
	for _, l := range leases {
		if opts.PathPrefix != "" && !strings.HasPrefix(l.Path, opts.PathPrefix) {
			continue
		}
		if opts.Severity != "" && !severityAtLeast(l.Severity, opts.Severity) {
			continue
		}
		result = append(result, l)
	}
	return result
}

// severityAtLeast returns true when actual >= minimum in rank.
func severityAtLeast(actual, minimum string) bool {
	return rank(actual) >= rank(minimum)
}

func rank(s string) int {
	switch strings.ToLower(s) {
	case "critical":
		return 2
	case "warning":
		return 1
	default:
		return 0
	}
}
