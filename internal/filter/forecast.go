package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/your-org/vaultpulse/internal/vault"
)

// ForecastEntry represents a predicted expiry event for a lease.
type ForecastEntry struct {
	Lease      vault.SecretLease
	ExpiresAt  time.Time
	TimeToLive time.Duration
	Severity   string
}

// ForecastOptions controls how far ahead to forecast and minimum severity.
type ForecastOptions struct {
	Window      time.Duration // how far into the future to look
	MinSeverity string        // "ok", "warn", "critical"
}

// DefaultForecastOptions returns sensible defaults.
func DefaultForecastOptions() ForecastOptions {
	return ForecastOptions{
		Window:      72 * time.Hour,
		MinSeverity: "warn",
	}
}

// Forecast returns leases predicted to expire within the given window,
// ordered by soonest expiry first.
func Forecast(leases []vault.SecretLease, opts ForecastOptions) []ForecastEntry {
	if opts.Window <= 0 {
		opts.Window = DefaultForecastOptions().Window
	}
	minRank := rank(opts.MinSeverity)
	now := time.Now()
	cutoff := now.Add(opts.Window)

	var entries []ForecastEntry
	for _, l := range leases {
		exp := l.ExpiresAt()
		if exp.IsZero() || exp.Before(now) || exp.After(cutoff) {
			continue
		}
		if rank(l.Severity) < minRank {
			continue
		}
		entries = append(entries, ForecastEntry{
			Lease:      l,
			ExpiresAt:  exp,
			TimeToLive: exp.Sub(now),
			Severity:   l.Severity,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ExpiresAt.Before(entries[j].ExpiresAt)
	})
	return entries
}

// PrintForecast writes a human-readable forecast table to w.
func PrintForecast(entries []ForecastEntry, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	if len(entries) == 0 {
		fmt.Fprintln(w, "No leases forecast to expire within the window.")
		return
	}
	fmt.Fprintf(w, "%-45s %-10s %-12s %s\n", "PATH", "SEVERITY", "TTL", "EXPIRES AT")
	fmt.Fprintf(w, "%s\n", "---------------------------------------------------------------------------------------------")
	for _, e := range entries {
		fmt.Fprintf(w, "%-45s %-10s %-12s %s\n",
			e.Lease.Path,
			e.Severity,
			formatDuration(e.TimeToLive),
			e.ExpiresAt.Format(time.RFC3339),
		)
	}
}
