// Package filter provides utilities for filtering and sorting secret leases.
package filter

import (
	"sort"

	"github.com/your-org/vaultpulse/internal/vault"
)

// SortField defines the field to sort leases by.
type SortField string

const (
	SortByExpiry   SortField = "expiry"
	SortBySeverity SortField = "severity"
	SortByPath     SortField = "path"
)

// SortOrder defines ascending or descending order.
type SortOrder string

const (
	Ascending  SortOrder = "asc"
	Descending SortOrder = "desc"
)

// SortOptions configures how leases are sorted.
type SortOptions struct {
	Field SortField
	Order SortOrder
}

// Sort returns a new slice of leases sorted according to opts.
// If opts.Field is empty, the original order is preserved.
func Sort(leases []vault.SecretLease, opts SortOptions) []vault.SecretLease {
	if len(leases) == 0 || opts.Field == "" {
		return leases
	}

	out := make([]vault.SecretLease, len(leases))
	copy(out, leases)

	sort.SliceStable(out, func(i, j int) bool {
		var less bool
		switch opts.Field {
		case SortByExpiry:
			less = out[i].ExpiresAt.Before(out[j].ExpiresAt)
		case SortBySeverity:
			less = rank(out[i].Severity) > rank(out[j].Severity)
		case SortByPath:
			less = out[i].Path < out[j].Path
		default:
			less = false
		}
		if opts.Order == Descending {
			return !less
		}
		return less
	})

	return out
}
