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

// DigestOptions controls what is included in the digest summary.
type DigestOptions struct {
	TopN      int    // number of top expiring leases to include
	MinSev    string // minimum severity to include
	Writer    io.Writer
}

// DefaultDigestOptions returns sensible defaults.
func DefaultDigestOptions() DigestOptions {
	return DigestOptions{
		TopN:   10,
		MinSev: "warn",
		Writer: os.Stdout,
	}
}

// DigestEntry is a single row in the digest.
type DigestEntry struct {
	LeaseID  string
	Path     string
	Severity string
	ExpiresIn time.Duration
}

// Digest builds a compact summary of the most urgent leases.
func Digest(leases []vault.SecretLease, opts DigestOptions) []DigestEntry {
	if opts.TopN <= 0 {
		opts.TopN = 10
	}

	filtered := Apply(leases, Options{MinSeverity: opts.MinSev})

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ExpiresAt.Before(filtered[j].ExpiresAt)
	})

	if len(filtered) > opts.TopN {
		filtered = filtered[:opts.TopN]
	}

	now := time.Now()
	entries := make([]DigestEntry, 0, len(filtered))
	for _, l := range filtered {
		entries = append(entries, DigestEntry{
			LeaseID:   l.LeaseID,
			Path:      l.Path,
			Severity:  l.Severity,
			ExpiresIn: l.ExpiresAt.Sub(now).Truncate(time.Second),
		})
	}
	return entries
}

// PrintDigest writes the digest table to the configured writer.
func PrintDigest(entries []DigestEntry, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	if len(entries) == 0 {
		fmt.Fprintln(w, "No leases match the digest criteria.")
		return
	}
	fmt.Fprintf(w, "%-36s  %-30s  %-8s  %s\n", "LEASE ID", "PATH", "SEVERITY", "EXPIRES IN")
	fmt.Fprintln(w, strings.Repeat("-", 90))
	for _, e := range entries {
		fmt.Fprintf(w, "%-36s  %-30s  %-8s  %s\n",
			truncate(e.LeaseID, 36),
			truncate(e.Path, 30),
			e.Severity,
			e.ExpiresIn,
		)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
