package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// ShadowOptions controls shadow lease detection behaviour.
type ShadowOptions struct {
	// MaxOverlap is the minimum TTL overlap (in seconds) between two leases
	// on the same path to be considered a shadow pair.
	MaxOverlap time.Duration
	// Out is where the report is written; defaults to os.Stdout.
	Out io.Writer
}

// DefaultShadowOptions returns sensible defaults.
func DefaultShadowOptions() ShadowOptions {
	return ShadowOptions{
		MaxOverlap: 5 * time.Minute,
		Out:        os.Stdout,
	}
}

// ShadowPair represents two leases on the same path whose validity windows
// overlap, meaning the newer lease "shadows" the older one.
type ShadowPair struct {
	Older   vault.SecretLease
	Newer   vault.SecretLease
	Overlap time.Duration
}

// DetectShadows finds leases that share a path and have overlapping TTL
// windows, returning pairs sorted by overlap duration descending.
func DetectShadows(leases []vault.SecretLease, opts ShadowOptions) []ShadowPair {
	if len(leases) == 0 {
		return nil
	}

	// Group by path.
	byPath := make(map[string][]vault.SecretLease)
	for _, l := range leases {
		byPath[l.Path] = append(byPath[l.Path], l)
	}

	var pairs []ShadowPair
	for _, group := range byPath {
		if len(group) < 2 {
			continue
		}
		// Sort by expiry ascending so older comes first.
		sort.Slice(group, func(i, j int) bool {
			return group[i].ExpiresAt.Before(group[j].ExpiresAt)
		})
		for i := 0; i < len(group)-1; i++ {
			for j := i + 1; j < len(group); j++ {
				older, newer := group[i], group[j]
				// Overlap: older must still be alive when newer was issued.
				if newer.IssuedAt.IsZero() || older.ExpiresAt.IsZero() {
					continue
				}
				overlap := older.ExpiresAt.Sub(newer.IssuedAt)
				if overlap >= opts.MaxOverlap {
					pairs = append(pairs, ShadowPair{
						Older:   older,
						Newer:   newer,
						Overlap: overlap,
					})
				}
			}
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Overlap > pairs[j].Overlap
	})
	return pairs
}

// PrintShadows writes a human-readable table of shadow pairs to opts.Out.
func PrintShadows(pairs []ShadowPair, opts ShadowOptions) {
	out := opts.Out
	if out == nil {
		out = os.Stdout
	}
	if len(pairs) == 0 {
		fmt.Fprintln(out, "No shadow leases detected.")
		return
	}
	fmt.Fprintf(out, "%-40s %-26s %-26s %s\n",
		"PATH", "OLDER EXPIRES", "NEWER ISSUED", "OVERLAP")
	fmt.Fprintln(out, strings.Repeat("-", 100))
	for _, p := range pairs {
		fmt.Fprintf(out, "%-40s %-26s %-26s %s\n",
			p.Older.Path,
			p.Older.ExpiresAt.Format(time.RFC3339),
			p.Newer.IssuedAt.Format(time.RFC3339),
			p.Overlap.Round(time.Second),
		)
	}
}
