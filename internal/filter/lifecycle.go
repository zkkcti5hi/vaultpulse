package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// LifecycleStage represents where a lease is in its lifecycle.
type LifecycleStage string

const (
	StageActive   LifecycleStage = "active"
	StageExpiring LifecycleStage = "expiring"
	StageExpired  LifecycleStage = "expired"
	StageRenewing LifecycleStage = "renewing"
)

// LifecycleEntry pairs a lease with its computed stage.
type LifecycleEntry struct {
	Lease vault.SecretLease
	Stage LifecycleStage
	Age   time.Duration
}

// ClassifyLifecycle assigns a LifecycleStage to each lease.
// Leases expiring within warnWindow are "expiring"; already past are "expired";
// those with RenewedAt set are "renewing"; otherwise "active".
func ClassifyLifecycle(leases []vault.SecretLease, warnWindow time.Duration) []LifecycleEntry {
	now := time.Now()
	entries := make([]LifecycleEntry, 0, len(leases))
	for _, l := range leases {
		var stage LifecycleStage
		switch {
		case l.ExpiresAt.Before(now):
			stage = StageExpired
		case l.ExpiresAt.Before(now.Add(warnWindow)):
			stage = StageExpiring
		case l.Metadata["renewed_at"] != "":
			stage = StageRenewing
		default:
			stage = StageActive
		}
		entries = append(entries, LifecycleEntry{
			Lease: l,
			Stage: stage,
			Age:   now.Sub(l.IssuedAt),
		})
	}
	return entries
}

// FilterByStage returns only entries matching the given stage.
func FilterByStage(entries []LifecycleEntry, stage LifecycleStage) []LifecycleEntry {
	out := entries[:0:0]
	for _, e := range entries {
		if e.Stage == stage {
			out = append(out, e)
		}
	}
	return out
}

// PrintLifecycle writes a formatted lifecycle table to w (defaults to os.Stdout).
func PrintLifecycle(entries []LifecycleEntry, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Lease.ExpiresAt.Before(entries[j].Lease.ExpiresAt)
	})
	fmt.Fprintf(w, "%-44s %-10s %-12s %s\n", "LEASE ID", "STAGE", "AGE", "EXPIRES AT")
	fmt.Fprintf(w, "%s\n", "-----------------------------------------------------------------------")
	for _, e := range entries {
		fmt.Fprintf(w, "%-44s %-10s %-12s %s\n",
			e.Lease.LeaseID,
			e.Stage,
			formatDuration(e.Age),
			e.Lease.ExpiresAt.Format(time.RFC3339),
		)
	}
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", h, m)
}
