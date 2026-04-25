package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/user/vaultpulse/internal/vault"
)

// AnomalyResult holds a lease flagged as anomalous along with a reason.
type AnomalyResult struct {
	Lease  vault.SecretLease
	Reason string
}

// AnomalyOptions controls anomaly detection behaviour.
type AnomalyOptions struct {
	// ShortTTLThreshold flags leases whose TTL is below this duration.
	ShortTTLThreshold time.Duration
	// RecentlySeenWindow flags leases first seen within this window.
	RecentlySeenWindow time.Duration
}

// DefaultAnomalyOptions returns sensible defaults.
func DefaultAnomalyOptions() AnomalyOptions {
	return AnomalyOptions{
		ShortTTLThreshold:  5 * time.Minute,
		RecentlySeenWindow: 10 * time.Minute,
	}
}

// DetectAnomalies inspects leases and returns those that appear anomalous.
func DetectAnomalies(leases []vault.SecretLease, opts AnomalyOptions) []AnomalyResult {
	now := time.Now()
	var results []AnomalyResult

	for _, l := range leases {
		ttl := time.Until(l.ExpiresAt)
		if ttl > 0 && ttl < opts.ShortTTLThreshold {
			results = append(results, AnomalyResult{
				Lease:  l,
				Reason: fmt.Sprintf("short TTL: %s remaining", ttl.Round(time.Second)),
			})
			continue
		}
		if !l.SeenAt.IsZero() && now.Sub(l.SeenAt) < opts.RecentlySeenWindow {
			results = append(results, AnomalyResult{
				Lease:  l,
				Reason: fmt.Sprintf("recently appeared: seen %s ago", now.Sub(l.SeenAt).Round(time.Second)),
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Lease.ExpiresAt.Before(results[j].Lease.ExpiresAt)
	})
	return results
}

// PrintAnomalies writes a human-readable anomaly report to w (defaults to os.Stdout).
func PrintAnomalies(results []AnomalyResult, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	if len(results) == 0 {
		fmt.Fprintln(w, "No anomalies detected.")
		return
	}
	fmt.Fprintf(w, "%-40s %-12s %s\n", "PATH", "SEVERITY", "REASON")
	fmt.Fprintf(w, "%-40s %-12s %s\n", "----", "--------", "------")
	for _, r := range results {
		fmt.Fprintf(w, "%-40s %-12s %s\n", r.Lease.Path, r.Lease.Severity, r.Reason)
	}
}
