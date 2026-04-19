package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/your-org/vaultpulse/internal/vault"
)

// ExpireWindow holds leases expiring within a given duration window.
type ExpireWindow struct {
	Window  time.Duration
	Leases  []vault.SecretLease
}

// FilterByExpireWindow returns leases expiring within the given duration from now.
func FilterByExpireWindow(leases []vault.SecretLease, window time.Duration) []vault.SecretLease {
	now := time.Now()
	cutoff := now.Add(window)
	out := make([]vault.SecretLease, 0, len(leases))
	for _, l := range leases {
		if l.ExpiresAt.After(now) && l.ExpiresAt.Before(cutoff) {
			out = append(out, l)
		}
	}
	return out
}

// GroupByExpireWindow buckets leases into named time windows.
func GroupByExpireWindow(leases []vault.SecretLease, windows map[string]time.Duration) map[string][]vault.SecretLease {
	result := make(map[string][]vault.SecretLease, len(windows))
	for name := range windows {
		result[name] = FilterByExpireWindow(leases, windows[name])
	}
	return result
}

// PrintExpireWindows writes a tabular summary of windowed expiry groups to w.
func PrintExpireWindows(w io.Writer, groups map[string][]vault.SecretLease, windowOrder []string) {
	if w == nil {
		w = os.Stdout
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "WINDOW\tCOUNT\tPATHS")
	for _, name := range windowOrder {
		leases := groups[name]
		paths := make([]string, 0, len(leases))
		for _, l := range leases {
			paths = append(paths, l.Path)
		}
		sort.Strings(paths)
		preview := ""
		if len(paths) > 0 {
			preview = paths[0]
			if len(paths) > 1 {
				preview += fmt.Sprintf(" (+%d more)", len(paths)-1)
			}
		}
		fmt.Fprintf(tw, "%s\t%d\t%s\n", name, len(leases), preview)
	}
	tw.Flush()
}
