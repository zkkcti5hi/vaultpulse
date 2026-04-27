package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// StaleOptions configures the stale lease detection behaviour.
type StaleOptions struct {
	// MinAge is the minimum time since SeenAt before a lease is considered stale.
	MinAge time.Duration
	// MaxResults caps the number of returned leases (0 = unlimited).
	MaxResults int
}

// DefaultStaleOptions returns sensible defaults: leases unseen for >24 h are stale.
func DefaultStaleOptions() StaleOptions {
	return StaleOptions{
		MinAge:     24 * time.Hour,
		MaxResults: 50,
	}
}

// DetectStale returns leases whose SeenAt timestamp is older than opts.MinAge,
// sorted by staleness (oldest first).
func DetectStale(leases []vault.SecretLease, opts StaleOptions) []vault.SecretLease {
	if len(leases) == 0 {
		return nil
	}
	now := time.Now()
	var stale []vault.SecretLease
	for _, l := range leases {
		seenAt := seenAtTime(l)
		if seenAt.IsZero() {
			continue
		}
		if now.Sub(seenAt) >= opts.MinAge {
			stale = append(stale, l)
		}
	}
	sort.Slice(stale, func(i, j int) bool {
		si := seenAtTime(stale[i])
		sj := seenAtTime(stale[j])
		return si.Before(sj)
	})
	if opts.MaxResults > 0 && len(stale) > opts.MaxResults {
		stale = stale[:opts.MaxResults]
	}
	return stale
}

// PrintStale writes a human-readable stale lease table to w.
func PrintStale(leases []vault.SecretLease, opts StaleOptions, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	if len(leases) == 0 {
		fmt.Fprintln(w, "No stale leases detected.")
		return
	}
	now := time.Now()
	fmt.Fprintf(w, "%-45s %-10s %s\n", "PATH", "SEVERITY", "LAST SEEN")
	fmt.Fprintf(w, "%-45s %-10s %s\n", "----", "--------", "---------")
	for _, l := range leases {
		age := "-"
		if t := seenAtTime(l); !t.IsZero() {
			age = fmt.Sprintf("%s ago", now.Sub(t).Truncate(time.Minute))
		}
		fmt.Fprintf(w, "%-45s %-10s %s\n", l.Path, l.Severity, age)
	}
}

// seenAtTime extracts the SeenAt value from a lease's metadata, returning zero
// time if absent or unparseable.
func seenAtTime(l vault.SecretLease) time.Time {
	raw, ok := l.Metadata["seen_at"]
	if !ok {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}
	}
	return t
}
