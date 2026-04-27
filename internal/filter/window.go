package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/vaultpulse/internal/vault"
)

// WindowOptions configures the rolling time window aggregation.
type WindowOptions struct {
	// WindowSize is the duration of each bucket (default: 1 hour).
	WindowSize time.Duration
	// NumWindows is the number of buckets to generate (default: 6).
	NumWindows int
	// Out is the writer for PrintWindows output (default: os.Stdout).
	Out io.Writer
}

// DefaultWindowOptions returns sensible defaults.
func DefaultWindowOptions() WindowOptions {
	return WindowOptions{
		WindowSize: time.Hour,
		NumWindows: 6,
		Out:        os.Stdout,
	}
}

// WindowBucket holds leases expiring within a specific time window.
type WindowBucket struct {
	Label  string
	Start  time.Time
	End    time.Time
	Leases []vault.SecretLease
}

// RollingWindow partitions leases into sequential time buckets starting from now.
func RollingWindow(leases []vault.SecretLease, opts WindowOptions) []WindowBucket {
	if opts.WindowSize <= 0 {
		opts.WindowSize = time.Hour
	}
	if opts.NumWindows <= 0 {
		opts.NumWindows = 6
	}

	now := time.Now()
	buckets := make([]WindowBucket, opts.NumWindows)
	for i := 0; i < opts.NumWindows; i++ {
		start := now.Add(time.Duration(i) * opts.WindowSize)
		end := start.Add(opts.WindowSize)
		buckets[i] = WindowBucket{
			Label: fmt.Sprintf("+%s", opts.WindowSize*time.Duration(i+1)),
			Start: start,
			End:   end,
		}
	}

	for _, l := range leases {
		exp := l.ExpiresAt
		if exp.Before(now) {
			continue
		}
		for i := range buckets {
			if !exp.Before(buckets[i].Start) && exp.Before(buckets[i].End) {
				buckets[i].Leases = append(buckets[i].Leases, l)
				break
			}
		}
	}
	return buckets
}

// PrintWindows writes a formatted rolling-window table to opts.Out.
func PrintWindows(buckets []WindowBucket, opts WindowOptions) {
	out := opts.Out
	if out == nil {
		out = os.Stdout
	}
	fmt.Fprintf(out, "%-12s  %-8s  %s\n", "WINDOW", "COUNT", "PATHS")
	fmt.Fprintf(out, "%-12s  %-8s  %s\n", "------", "-----", "-----")
	for _, b := range buckets {
		paths := uniquePaths(b.Leases)
		sort.Strings(paths)
		preview := ""
		if len(paths) > 0 {
			preview = paths[0]
			if len(paths) > 1 {
				preview += fmt.Sprintf(" (+%d)", len(paths)-1)
			}
		}
		fmt.Fprintf(out, "%-12s  %-8d  %s\n", b.Label, len(b.Leases), preview)
	}
}

func uniquePaths(leases []vault.SecretLease) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, l := range leases {
		if _, ok := seen[l.Path]; !ok {
			seen[l.Path] = struct{}{}
			out = append(out, l.Path)
		}
	}
	return out
}
