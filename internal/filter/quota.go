package filter

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// QuotaOptions controls how lease quota enforcement is evaluated.
type QuotaOptions struct {
	// MaxPerPath is the maximum number of active leases allowed per path prefix.
	MaxPerPath int
	// MaxPerSeverity is the maximum number of leases allowed per severity level.
	MaxPerSeverity int
}

// DefaultQuotaOptions returns sensible defaults for quota enforcement.
func DefaultQuotaOptions() QuotaOptions {
	return QuotaOptions{
		MaxPerPath:     20,
		MaxPerSeverity: 50,
	}
}

// QuotaViolation describes a single quota breach.
type QuotaViolation struct {
	Dimension string // "path" or "severity"
	Key       string
	Count     int
	Limit     int
}

func (v QuotaViolation) String() string {
	return fmt.Sprintf("%s=%q: %d leases exceeds limit of %d", v.Dimension, v.Key, v.Count, v.Limit)
}

// ApplyQuota checks leases against configured limits and returns any violations.
func ApplyQuota(leases []vault.SecretLease, opts QuotaOptions) []QuotaViolation {
	var violations []QuotaViolation

	pathCounts := make(map[string]int)
	for _, l := range leases {
		pathCounts[pathPrefix(l.Path)]++
	}
	pathKeys := make([]string, 0, len(pathCounts))
	for k := range pathCounts {
		pathKeys = append(pathKeys, k)
	}
	sort.Strings(pathKeys)
	for _, k := range pathKeys {
		if c := pathCounts[k]; opts.MaxPerPath > 0 && c > opts.MaxPerPath {
			violations = append(violations, QuotaViolation{Dimension: "path", Key: k, Count: c, Limit: opts.MaxPerPath})
		}
	}

	sevCounts := make(map[string]int)
	for _, l := range leases {
		sevCounts[l.Severity]++
	}
	sevKeys := make([]string, 0, len(sevCounts))
	for k := range sevCounts {
		sevKeys = append(sevKeys, k)
	}
	sort.Strings(sevKeys)
	for _, k := range sevKeys {
		if c := sevCounts[k]; opts.MaxPerSeverity > 0 && c > opts.MaxPerSeverity {
			violations = append(violations, QuotaViolation{Dimension: "severity", Key: k, Count: c, Limit: opts.MaxPerSeverity})
		}
	}

	return violations
}

// PrintQuota writes a human-readable quota report to w.
func PrintQuota(violations []QuotaViolation, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	if len(violations) == 0 {
		fmt.Fprintln(w, "No quota violations detected.")
		return
	}
	fmt.Fprintf(w, "%-12s %-30s %8s %8s\n", "DIMENSION", "KEY", "COUNT", "LIMIT")
	fmt.Fprintf(w, "%-12s %-30s %8s %8s\n", "----------", "----------------------------", "-------", "-------")
	for _, v := range violations {
		fmt.Fprintf(w, "%-12s %-30s %8d %8d\n", v.Dimension, v.Key, v.Count, v.Limit)
	}
}
