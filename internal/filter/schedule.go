package filter

import (
	"fmt"
	"time"

	"github.com/your-org/vaultpulse/internal/vault"
)

// ScheduleEntry represents a scheduled alert window for a lease.
type ScheduleEntry struct {
	LeaseID  string
	Path     string
	NotifyAt time.Time
	Severity string
}

// String returns a human-readable representation of the schedule entry.
func (s ScheduleEntry) String() string {
	return fmt.Sprintf("[%s] %s — notify at %s (severity: %s)",
		s.LeaseID, s.Path, s.NotifyAt.Format(time.RFC3339), s.Severity)
}

// BuildSchedule returns a list of ScheduleEntries for leases that should be
// alerted within the given lookahead window from now.
func BuildSchedule(leases []vault.SecretLease, lookahead time.Duration) []ScheduleEntry {
	now := time.Now()
	cutoff := now.Add(lookahead)
	var entries []ScheduleEntry
	for _, l := range leases {
		if l.ExpiresAt.IsZero() {
			continue
		}
		if l.ExpiresAt.After(now) && l.ExpiresAt.Before(cutoff) {
			entries = append(entries, ScheduleEntry{
				LeaseID:  l.LeaseID,
				Path:     l.Path,
				NotifyAt: l.ExpiresAt.Add(-lookahead / 2),
				Severity: l.Severity,
			})
		}
	}
	return entries
}

// FilterScheduleByMinSeverity removes entries below the given minimum severity.
func FilterScheduleByMinSeverity(entries []ScheduleEntry, minSeverity string) []ScheduleEntry {
	var out []ScheduleEntry
	for _, e := range entries {
		if rank(e.Severity) >= rank(minSeverity) {
			out = append(out, e)
		}
	}
	return out
}
