package filter

import (
	"strings"
	"unicode"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// NormalizeOptions controls how lease fields are normalized.
type NormalizeOptions struct {
	// TrimSpace removes leading/trailing whitespace from string fields.
	TrimSpace bool
	// LowercasePath converts the lease Path to lowercase.
	LowercasePath bool
	// LowercaseLeaseID converts the LeaseID to lowercase.
	LowercaseLeaseID bool
	// CollapseMetaSpaces replaces runs of whitespace in metadata values with a single space.
	CollapseMetaSpaces bool
}

// DefaultNormalizeOptions returns a NormalizeOptions with sensible defaults.
func DefaultNormalizeOptions() NormalizeOptions {
	return NormalizeOptions{
		TrimSpace:          true,
		LowercasePath:      false,
		LowercaseLeaseID:   false,
		CollapseMetaSpaces: true,
	}
}

// Normalize returns a new slice of leases with string fields cleaned up
// according to opts. The original slice is never mutated.
func Normalize(leases []vault.SecretLease, opts NormalizeOptions) []vault.SecretLease {
	out := make([]vault.SecretLease, 0, len(leases))
	for _, l := range leases {
		out = append(out, normalizeOne(l, opts))
	}
	return out
}

func normalizeOne(l vault.SecretLease, opts NormalizeOptions) vault.SecretLease {
	if opts.TrimSpace {
		l.Path = strings.TrimSpace(l.Path)
		l.LeaseID = strings.TrimSpace(l.LeaseID)
	}
	if opts.LowercasePath {
		l.Path = strings.ToLower(l.Path)
	}
	if opts.LowercaseLeaseID {
		l.LeaseID = strings.ToLower(l.LeaseID)
	}
	if opts.CollapseMetaSpaces && l.Metadata != nil {
		newMeta := make(map[string]string, len(l.Metadata))
		for k, v := range l.Metadata {
			newMeta[k] = collapseSpaces(v)
		}
		l.Metadata = newMeta
	}
	return l
}

func collapseSpaces(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !inSpace {
				b.WriteRune(' ')
				inSpace = true
			}
		} else {
			b.WriteRune(r)
			inSpace = false
		}
	}
	return strings.TrimSpace(b.String())
}
