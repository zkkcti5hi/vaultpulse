package filter

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// PatternOptions configures pattern-based lease matching.
type PatternOptions struct {
	// Patterns is a list of glob-style path patterns (e.g. "secret/db/*").
	Patterns []string
	// Invert returns leases that do NOT match any pattern when true.
	Invert bool
}

// DefaultPatternOptions returns PatternOptions with sensible defaults.
func DefaultPatternOptions() PatternOptions {
	return PatternOptions{
		Patterns: nil,
		Invert:   false,
	}
}

// MatchPattern reports whether path matches a simple glob pattern.
// Only '*' wildcards are supported (matches any sequence of non-slash chars).
func MatchPattern(pattern, path string) bool {
	if pattern == "*" {
		return true
	}
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == path
	}
	remaining := path
	for i, part := range parts {
		if part == "" {
			continue
		}
		idx := strings.Index(remaining, part)
		if idx == -1 {
			return false
		}
		if i == 0 && !strings.HasPrefix(path, part) {
			return false
		}
		remaining = remaining[idx+len(part):]
	}
	if !strings.HasSuffix(pattern, "*") && remaining != "" {
		return false
	}
	return true
}

// FilterByPattern returns leases whose paths match at least one of the given
// patterns. If opts.Invert is true, leases that do NOT match are returned.
func FilterByPattern(leases []vault.SecretLease, opts PatternOptions) []vault.SecretLease {
	if len(opts.Patterns) == 0 {
		return leases
	}
	out := make([]vault.SecretLease, 0, len(leases))
	for _, l := range leases {
		matched := false
		for _, p := range opts.Patterns {
			if MatchPattern(p, l.Path) {
				matched = true
				break
			}
		}
		if matched != opts.Invert {
			out = append(out, l)
		}
	}
	return out
}

// PrintPattern writes a summary of matched leases to w.
func PrintPattern(leases []vault.SecretLease, opts PatternOptions, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, "%-40s  %-10s  %s\n", "PATH", "SEVERITY", "LEASE ID")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 72))
	for _, l := range leases {
		fmt.Fprintf(w, "%-40s  %-10s  %s\n", l.Path, l.Severity, l.LeaseID)
	}
	fmt.Fprintf(w, "\n%d lease(s) matched pattern(s): %s\n", len(leases), strings.Join(opts.Patterns, ", "))
}
