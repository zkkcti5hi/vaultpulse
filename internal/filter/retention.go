package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// RetentionPolicy defines how long leases should be retained before flagging.
type RetentionPolicy struct {
	// MaxAge is the maximum allowed age of a lease (time since SeenAt).
	MaxAge time.Duration
	// MaxTTL is the maximum allowed total TTL for a lease.
	MaxTTL time.Duration
}

// RetentionResult holds a lease that violates a retention policy.
type RetentionResult struct {
	Lease   vault.SecretLease
	Reason  string
	Age     time.Duration
}

// DefaultRetentionOptions returns sensible defaults.
func DefaultRetentionOptions() RetentionPolicy {
	return RetentionPolicy{
		MaxAge: 30 * 24 * time.Hour, // 30 days
		MaxTTL: 90 * 24 * time.Hour, // 90 days
	}
}

// ApplyRetention flags leases that violate the given retention policy.
// A lease is flagged if its age exceeds MaxAge or its TTL exceeds MaxTTL.
func ApplyRetention(leases []vault.SecretLease, policy RetentionPolicy) []RetentionResult {
	now := time.Now()
	var results []RetentionResult

	for _, l := range leases {
		var reason string
		var age time.Duration

		if !l.SeenAt.IsZero() {
			age = now.Sub(l.SeenAt)
			if policy.MaxAge > 0 && age > policy.MaxAge {
				reason = fmt.Sprintf("age %s exceeds max %s", fmtDur(age), fmtDur(policy.MaxAge))
			}
		}

		ttl := time.Duration(l.TTL) * time.Second
		if reason == "" && policy.MaxTTL > 0 && ttl > policy.MaxTTL {
			reason = fmt.Sprintf("TTL %s exceeds max %s", fmtDur(ttl), fmtDur(policy.MaxTTL))
		}

		if reason != "" {
			results = append(results, RetentionResult{
				Lease:  l,
				Reason: reason,
				Age:    age,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Lease.LeaseID < results[j].Lease.LeaseID
	})
	return results
}

// PrintRetention writes a human-readable retention report to w.
func PrintRetention(results []RetentionResult, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	if len(results) == 0 {
		fmt.Fprintln(w, "No retention violations found.")
		return
	}
	fmt.Fprintf(w, "%-40s %-30s %s\n", "LEASE ID", "PATH", "REASON")
	fmt.Fprintf(w, "%-40s %-30s %s\n", "--------", "----", "------")
	for _, r := range results {
		fmt.Fprintf(w, "%-40s %-30s %s\n", r.Lease.LeaseID, r.Lease.Path, r.Reason)
	}
}

func fmtDur(d time.Duration) string {
	days := int(d.Hours()) / 24
	if days > 0 {
		return fmt.Sprintf("%dd", days)
	}
	return d.Round(time.Second).String()
}
