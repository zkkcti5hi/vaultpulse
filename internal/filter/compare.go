package filter

import (
	"fmt"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// CompareResult holds the diff between two lease snapshots.
type CompareResult struct {
	Added   []vault.Lease
	Removed []vault.Lease
	Changed []vault.Lease
}

// String returns a human-readable summary of the comparison.
func (r CompareResult) String() string {
	return fmt.Sprintf("Added: %d, Removed: %d, Changed: %d",
		len(r.Added), len(r.Removed), len(r.Changed))
}

// Compare diffs two slices of leases by LeaseID.
// A lease is "changed" if its Severity or TTL bucket differs between snapshots.
func Compare(before, after []vault.Lease) CompareResult {
	beforeMap := make(map[string]vault.Lease, len(before))
	for _, l := range before {
		beforeMap[l.LeaseID] = l
	}

	afterMap := make(map[string]vault.Lease, len(after))
	for _, l := range after {
		afterMap[l.LeaseID] = l
	}

	var result CompareResult

	for id, a := range afterMap {
		b, exists := beforeMap[id]
		if !exists {
			result.Added = append(result.Added, a)
		} else if b.Severity != a.Severity {
			result.Changed = append(result.Changed, a)
		}
	}

	for id, b := range beforeMap {
		if _, exists := afterMap[id]; !exists {
			result.Removed = append(result.Removed, b)
		}
	}

	return result
}
