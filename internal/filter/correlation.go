package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// CorrelationGroup holds leases that share a common attribute.
type CorrelationGroup struct {
	Key    string
	Leases []vault.SecretLease
}

// CorrelationResult is the output of a Correlate call.
type CorrelationResult struct {
	Field  string
	Groups []CorrelationGroup
}

// Correlate groups leases by a shared field ("path-prefix", "severity", "tag") and
// returns groups with more than one member so operators can spot related expirations.
func Correlate(leases []vault.SecretLease, field string) CorrelationResult {
	index := map[string][]vault.SecretLease{}
	for _, l := range leases {
		key := correlationKey(l, field)
		if key == "" {
			continue
		}
		index[key] = append(index[key], l)
	}

	var groups []CorrelationGroup
	for k, ls := range index {
		if len(ls) < 2 {
			continue
		}
		groups = append(groups, CorrelationGroup{Key: k, Leases: ls})
	}
	sort.Slice(groups, func(i, j int) bool {
		if len(groups[i].Leases) != len(groups[j].Leases) {
			return len(groups[i].Leases) > len(groups[j].Leases)
		}
		return groups[i].Key < groups[j].Key
	})
	return CorrelationResult{Field: field, Groups: groups}
}

func correlationKey(l vault.SecretLease, field string) string {
	switch field {
	case "severity":
		return l.Severity
	case "tag":
		if len(l.Tags) == 0 {
			return ""
		}
		return strings.Join(l.Tags, ",")
	default: // path-prefix
		return pathPrefix(l.Path)
	}
}

// PrintCorrelation writes a human-readable correlation table to w (defaults to stdout).
func PrintCorrelation(r CorrelationResult, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, "Correlation by %s\n", r.Field)
	fmt.Fprintf(w, "%-30s  %6s  %s\n", "Key", "Count", "Paths")
	fmt.Fprintln(w, strings.Repeat("-", 70))
	for _, g := range r.Groups {
		paths := make([]string, 0, len(g.Leases))
		for _, l := range g.Leases {
			paths = append(paths, l.Path)
		}
		fmt.Fprintf(w, "%-30s  %6d  %s\n", g.Key, len(g.Leases), strings.Join(paths, ", "))
	}
}
