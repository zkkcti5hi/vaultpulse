package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// RollupEntry aggregates lease statistics for a given grouping key.
type RollupEntry struct {
	Key          string
	Count        int
	Critical     int
	Warn         int
	OK           int
	EarliestExp  time.Time
	LatestExp    time.Time
}

// RollupOptions controls how Rollup groups and filters leases.
type RollupOptions struct {
	// GroupBy is one of "severity", "path", or "tag".
	GroupBy string
}

// DefaultRollupOptions returns sensible defaults.
func DefaultRollupOptions() RollupOptions {
	return RollupOptions{GroupBy: "severity"}
}

// Rollup aggregates leases into summary entries grouped by the chosen dimension.
func Rollup(leases []vault.SecretLease, opts RollupOptions) []RollupEntry {
	index := map[string]*RollupEntry{}

	for _, l := range leases {
		keys := rollupKeys(l, opts.GroupBy)
		for _, key := range keys {
			e, ok := index[key]
			if !ok {
				e = &RollupEntry{Key: key, EarliestExp: l.ExpiresAt, LatestExp: l.ExpiresAt}
				index[key] = e
			}
			e.Count++
			switch l.Severity {
			case "critical":
				e.Critical++
			case "warn":
				e.Warn++
			default:
				e.OK++
			}
			if l.ExpiresAt.Before(e.EarliestExp) {
				e.EarliestExp = l.ExpiresAt
			}
			if l.ExpiresAt.After(e.LatestExp) {
				e.LatestExp = l.ExpiresAt
			}
		}
	}

	result := make([]RollupEntry, 0, len(index))
	for _, e := range index {
		result = append(result, *e)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result
}

// PrintRollup writes a formatted rollup table to w (defaults to os.Stdout).
func PrintRollup(entries []RollupEntry, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, "%-30s %6s %8s %6s %4s  %s\n", "KEY", "COUNT", "CRITICAL", "WARN", "OK", "EARLIEST_EXPIRY")
	fmt.Fprintf(w, "%s\n", "----------------------------------------------------------------------")
	for _, e := range entries {
		fmt.Fprintf(w, "%-30s %6d %8d %6d %4d  %s\n",
			e.Key, e.Count, e.Critical, e.Warn, e.OK,
			e.EarliestExp.Format(time.RFC3339))
	}
}

func rollupKeys(l vault.SecretLease, groupBy string) []string {
	switch groupBy {
	case "path":
		return []string{pathPrefix(l.Path)}
	case "tag":
		if len(l.Tags) == 0 {
			return []string{"(untagged)"}
		}
		out := make([]string, len(l.Tags))
		copy(out, l.Tags)
		return out
	default: // "severity"
		if l.Severity == "" {
			return []string{"ok"}
		}
		return []string{l.Severity}
	}
}
