package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// VelocityOptions controls how lease expiry velocity is computed.
type VelocityOptions struct {
	// WindowSize is the duration over which expiry rate is measured.
	WindowSize time.Duration
	// MinLeases is the minimum number of leases required to emit a result.
	MinLeases int
}

// DefaultVelocityOptions returns sensible defaults.
func DefaultVelocityOptions() VelocityOptions {
	return VelocityOptions{
		WindowSize: 24 * time.Hour,
		MinLeases:  1,
	}
}

// VelocityResult holds the computed expiry rate for a path prefix.
type VelocityResult struct {
	Prefix      string
	Count       int
	WindowSize  time.Duration
	RatePerHour float64
	Severity    string
}

// Velocity computes the rate at which leases expire within the given window,
// grouped by path prefix.
func Velocity(leases []vault.SecretLease, opts VelocityOptions) []VelocityResult {
	now := time.Now()
	cutoff := now.Add(opts.WindowSize)

	counts := make(map[string]int)
	severities := make(map[string]string)

	for _, l := range leases {
		exp := l.ExpiresAt()
		if exp.Before(now) || exp.After(cutoff) {
			continue
		}
		pfx := pathPrefix(l.Path)
		counts[pfx]++
		if rank(l.Severity) > rank(severities[pfx]) {
			severities[pfx] = l.Severity
		}
	}

	hours := opts.WindowSize.Hours()
	if hours == 0 {
		hours = 1
	}

	var results []VelocityResult
	for pfx, count := range counts {
		if count < opts.MinLeases {
			continue
		}
		results = append(results, VelocityResult{
			Prefix:      pfx,
			Count:       count,
			WindowSize:  opts.WindowSize,
			RatePerHour: float64(count) / hours,
			Severity:    severities[pfx],
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Count != results[j].Count {
			return results[i].Count > results[j].Count
		}
		return results[i].Prefix < results[j].Prefix
	})

	return results
}

// PrintVelocity writes a formatted velocity table to w (defaults to os.Stdout).
func PrintVelocity(results []VelocityResult, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, "%-40s %8s %12s %10s\n", "PREFIX", "COUNT", "RATE/HR", "SEVERITY")
	fmt.Fprintf(w, "%s\n", fmt.Sprintf("%0*d", 74, 0)[:74])
	if len(results) == 0 {
		fmt.Fprintln(w, "no velocity data")
		return
	}
	for _, r := range results {
		fmt.Fprintf(w, "%-40s %8d %12.4f %10s\n", r.Prefix, r.Count, r.RatePerHour, r.Severity)
	}
}
