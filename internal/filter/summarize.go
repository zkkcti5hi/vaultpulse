package filter

import (
	"fmt"

	"github.com/your-org/vaultpulse/internal/vault"
)

// Summary holds aggregated statistics about a set of leases.
type Summary struct {
	Total    int
	BySeverity map[string]int
	ExpiredCount int
	CriticalPaths []string
}

// Summarize computes a Summary from a slice of annotated leases.
func Summarize(leases []vault.SecretLease) Summary {
	s := Summary{
		Total:      len(leases),
		BySeverity: make(map[string]int),
	}

	seen := map[string]bool{}
	for _, l := range leases {
		sev := l.Severity
		if sev == "" {
			sev = "ok"
		}
		s.BySeverity[sev]++

		if l.TTL <= 0 {
			s.ExpiredCount++
		}

		if sev == "critical" {
			prefix := pathPrefix(l.Path)
			if !seen[prefix] {
				seen[prefix] = true
				s.CriticalPaths = append(s.CriticalPaths, prefix)
			}
		}
	}
	return s
}

// String returns a human-readable one-line summary.
func (s Summary) String() string {
	return fmt.Sprintf(
		"total=%d expired=%d critical=%d warning=%d ok=%d",
		s.Total,
		s.ExpiredCount,
		s.BySeverity["critical"],
		s.BySeverity["warning"],
		s.BySeverity["ok"],
	)
}
