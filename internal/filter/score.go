package filter

import (
	"sort"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// LeaseScore holds a lease with its computed risk score.
type LeaseScore struct {
	Lease vault.SecretLease
	Score int
}

// Score computes a risk score for each lease based on severity and TTL.
// Critical = 100 base, Warning = 50, Info = 10.
// Each tag adds 5 points. Each label adds 3 points.
func Score(leases []vault.SecretLease) []LeaseScore {
	results := make([]LeaseScore, 0, len(leases))
	for _, l := range leases {
		results = append(results, LeaseScore{
			Lease: l,
			Score: computeScore(l),
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results
}

func computeScore(l vault.SecretLease) int {
	base := 0
	switch l.Severity {
	case "critical":
		base = 100
	case "warning":
		base = 50
	case "info":
		base = 10
	}
	base += len(l.Tags) * 5
	base += len(l.Labels) * 3
	return base
}
