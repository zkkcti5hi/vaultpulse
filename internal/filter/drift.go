package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// DriftOptions configures drift detection behaviour.
type DriftOptions struct {
	// BaselineTTL is the expected TTL for a healthy lease.
	BaselineTTL time.Duration
	// TolerancePct is the allowed deviation from BaselineTTL (0–1).
	TolerancePct float64
	// MinSeverity filters leases below this severity before analysis.
	MinSeverity string
}

// DefaultDriftOptions returns sensible defaults.
func DefaultDriftOptions() DriftOptions {
	return DriftOptions{
		BaselineTTL:  24 * time.Hour,
		TolerancePct: 0.20,
		MinSeverity:  "ok",
	}
}

// DriftResult holds a lease together with its measured drift.
type DriftResult struct {
	Lease     vault.SecretLease
	ActualTTL time.Duration
	DriftPct  float64 // positive = longer than baseline, negative = shorter
	Exceeds   bool    // true when abs(DriftPct) > TolerancePct
}

// DetectDrift compares each lease TTL against the baseline and returns
// results sorted by absolute drift descending.
func DetectDrift(leases []vault.SecretLease, opts DriftOptions) []DriftResult {
	if len(leases) == 0 {
		return nil
	}

	baseSecs := opts.BaselineTTL.Seconds()
	now := time.Now()

	var results []DriftResult
	for _, l := range leases {
		ttl := time.Until(l.ExpiresAt)
		if ttl < 0 {
			ttl = 0
		}
		_ = now
		var driftPct float64
		if baseSecs > 0 {
			driftPct = (ttl.Seconds() - baseSecs) / baseSecs
		}
		abs := driftPct
		if abs < 0 {
			abs = -abs
		}
		results = append(results, DriftResult{
			Lease:     l,
			ActualTTL: ttl,
			DriftPct:  driftPct,
			Exceeds:   abs > opts.TolerancePct,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		ai, aj := results[i].DriftPct, results[j].DriftPct
		if ai < 0 {
			ai = -ai
		}
		if aj < 0 {
			aj = -aj
		}
		return ai > aj
	})
	return results
}

// PrintDrift writes a human-readable drift table to w (defaults to os.Stdout).
func PrintDrift(results []DriftResult, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, "%-40s  %-12s  %-10s  %s\n", "LEASE PATH", "ACTUAL TTL", "DRIFT %", "EXCEEDS")
	fmt.Fprintf(w, "%s\n", "------------------------------------------------------------------------")
	for _, r := range results {
		exceeds := ""
		if r.Exceeds {
			exceeds = "YES"
		}
		fmt.Fprintf(w, "%-40s  %-12s  %+9.1f%%  %s\n",
			r.Lease.Path,
			r.ActualTTL.Round(time.Second).String(),
			r.DriftPct*100,
			exceeds,
		)
	}
}
