package filter

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/your-org/vaultpulse/internal/vault"
)

// DiffResult holds the output of comparing two snapshots.
type DiffResult struct {
	Added   []vault.SecretLease
	Removed []vault.SecretLease
	Changed []vault.SecretLease
}

// Summary returns a one-line summary of the diff.
func (d DiffResult) Summary() string {
	return fmt.Sprintf("added=%d removed=%d changed=%d",
		len(d.Added), len(d.Removed), len(d.Changed))
}

// IsEmpty returns true when there are no differences.
func (d DiffResult) IsEmpty() bool {
	return len(d.Added) == 0 && len(d.Removed) == 0 && len(d.Changed) == 0
}

// Diff compares two lease slices and returns a DiffResult.
func Diff(before, after []vault.SecretLease) DiffResult {
	beforeMap := make(map[string]vault.SecretLease, len(before))
	for _, l := range before {
		beforeMap[l.LeaseID] = l
	}
	afterMap := make(map[string]vault.SecretLease, len(after))
	for _, l := range after {
		afterMap[l.LeaseID] = l
	}

	var result DiffResult
	for id, a := range afterMap {
		if b, ok := beforeMap[id]; !ok {
			result.Added = append(result.Added, a)
		} else if b.TTL != a.TTL || b.Severity != a.Severity {
			result.Changed = append(result.Changed, a)
		}
	}
	for id, b := range beforeMap {
		if _, ok := afterMap[id]; !ok {
			result.Removed = append(result.Removed, b)
		}
	}
	return result
}

// PrintDiff writes a human-readable diff table to w.
func PrintDiff(w io.Writer, d DiffResult) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "STATUS\tPATH\tLEASE ID\tSEVERITY")
	for _, l := range d.Added {
		fmt.Fprintf(tw, "+ added\t%s\t%s\t%s\n", l.Path, l.LeaseID, l.Severity)
	}
	for _, l := range d.Removed {
		fmt.Fprintf(tw, "- removed\t%s\t%s\t%s\n", l.Path, l.LeaseID, l.Severity)
	}
	for _, l := range d.Changed {
		fmt.Fprintf(tw, "~ changed\t%s\t%s\t%s\n", l.Path, l.LeaseID, l.Severity)
	}
	tw.Flush()
}
