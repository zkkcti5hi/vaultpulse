package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// ReplayEvent represents a lease state captured at a point in time.
type ReplayEvent struct {
	At     time.Time
	Leases []vault.SecretLease
}

// ReplayStore holds an ordered sequence of replay events.
type ReplayStore struct {
	events []ReplayEvent
}

// NewReplayStore creates an empty ReplayStore.
func NewReplayStore() *ReplayStore {
	return &ReplayStore{}
}

// Record appends a snapshot of leases at the given time.
func (r *ReplayStore) Record(at time.Time, leases []vault.SecretLease) {
	copy := make([]vault.SecretLease, len(leases))
	for i, l := range leases {
		copy[i] = l
	}
	r.events = append(r.events, ReplayEvent{At: at, Leases: copy})
}

// Len returns the number of recorded events.
func (r *ReplayStore) Len() int {
	return len(r.events)
}

// At returns the event closest to the requested time, or false if empty.
func (r *ReplayStore) At(t time.Time) (ReplayEvent, bool) {
	if len(r.events) == 0 {
		return ReplayEvent{}, false
	}
	best := r.events[0]
	bestDiff := absDuration(t.Sub(best.At))
	for _, e := range r.events[1:] {
		d := absDuration(t.Sub(e.At))
		if d < bestDiff {
			best = e
			bestDiff = d
		}
	}
	return best, true
}

// All returns all events sorted ascending by time.
func (r *ReplayStore) All() []ReplayEvent {
	out := make([]ReplayEvent, len(r.events))
	copy(out, r.events)
	sort.Slice(out, func(i, j int) bool {
		return out[i].At.Before(out[j].At)
	})
	return out
}

// PrintReplay writes a human-readable replay timeline to w.
func PrintReplay(w io.Writer, events []ReplayEvent) {
	if w == nil {
		w = os.Stdout
	}
	if len(events) == 0 {
		fmt.Fprintln(w, "No replay events recorded.")
		return
	}
	fmt.Fprintf(w, "%-30s  %-6s  %s\n", "TIME", "COUNT", "TOP SEVERITY")
	fmt.Fprintf(w, "%-30s  %-6s  %s\n", "----", "-----", "------------")
	for _, e := range events {
		top := topSeverity(e.Leases)
		fmt.Fprintf(w, "%-30s  %-6d  %s\n", e.At.Format(time.RFC3339), len(e.Leases), top)
	}
}

func topSeverity(leases []vault.SecretLease) string {
	best := "ok"
	for _, l := range leases {
		if severityRank(l.Severity) > severityRank(best) {
			best = l.Severity
		}
	}
	return best
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
