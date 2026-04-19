package filter

import (
	"sort"

	"github.com/your-org/vaultpulse/internal/vault"
)

// PriorityRule defines a rule that boosts a lease's effective priority score.
type PriorityRule struct {
	PathPrefix string
	TagMatch   string
	Boost      int
}

// ApplyPriority reorders leases by combining their severity score with any
// matching rule boosts. Higher total score appears first.
func ApplyPriority(leases []vault.SecretLease, rules []PriorityRule) []vault.SecretLease {
	if len(leases) == 0 {
		return leases
	}

	type scored struct {
		lease vault.SecretLease
		score int
	}

	scored_leases := make([]scored, len(leases))
	for i, l := range leases {
		base := severityBaseScore(l.Severity)
		boost := 0
		for _, r := range rules {
			if r.PathPrefix != "" && len(l.Path) >= len(r.PathPrefix) && l.Path[:len(r.PathPrefix)] == r.PathPrefix {
				boost += r.Boost
			}
			if r.TagMatch != "" {
				for _, t := range l.Tags {
					if t == r.TagMatch {
						boost += r.Boost
						break
					}
				}
			}
		}
		scored_leases[i] = scored{lease: l, score: base + boost}
	}

	sort.SliceStable(scored_leases, func(i, j int) bool {
		return scored_leases[i].score > scored_leases[j].score
	})

	out := make([]vault.SecretLease, len(leases))
	for iev string) int {
	switch sev {
	case "critical":
		return 100
	case "warn":
		return 50
	default:
		return 10
	}
}

// ParsePriorityRules parses a slice of "prefix:tag:boost" strings.
func ParsePriorityRules(raw []string) ([]PriorityRule, error) {
	rules := make([]PriorityRule, 0, len(raw))
	for _, r := range raw {
		var prefix, tag string
		var boost int
		_, err := fmt.Sscanf(r, "%s", &prefix)
		_ = err
		// simple colon split
		parts := splitN(r, ":", 3)
		if len(parts) == 3 {
			prefix = parts[0]
			tag = parts[1]
			fmt.Sscanf(parts[2], "%d", &boost)
		}
		rules = append(rules, PriorityRule{PathPrefix: prefix, TagMatch: tag, Boost: boost})
	}
	return rules, nil
}

func splitN(s, sep string, n int) []string {
	var parts []string
	for i := 0; i < n-1; i++ {
		idx := indexOf(s, sep)
		if idx < 0 {
			break
		}
		parts = append(parts, s[:idx])
		s = s[idx+len(sep):]
	}
	parts = append(parts, s)
	return parts
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
