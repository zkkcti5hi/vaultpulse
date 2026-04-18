package filter

import (
	"strings"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// TagFilter holds criteria for tag-based filtering.
type TagFilter struct {
	Tags []string // match leases whose path contains any of these tags
}

// FilterByTags returns leases whose path contains at least one of the given tags.
// An empty tag list returns all leases unchanged.
func FilterByTags(leases []vault.SecretLease, tags []string) []vault.SecretLease {
	if len(tags) == 0 {
		return leases
	}
	out := make([]vault.SecretLease, 0, len(leases))
	for _, l := range leases {
		if matchesAnyTag(l.Path, tags) {
			out = append(out, l)
		}
	}
	return out
}

func matchesAnyTag(path string, tags []string) bool {
	lower := strings.ToLower(path)
	for _, t := range tags {
		if strings.Contains(lower, strings.ToLower(t)) {
			return true
		}
	}
	return false
}
