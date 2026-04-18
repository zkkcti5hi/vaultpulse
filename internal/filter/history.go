package filter

import (
	"sync"
	"time"

	"github.com/your-org/vaultpulse/internal/vault"
)

// SnapshotEntry records leases captured at a point in time.
type SnapshotEntry struct {
	CapturedAt time.Time
	Leases     []vault.SecretLease
}

// History stores a rolling window of lease snapshots.
type History struct {
	mu       sync.RWMutex
	max      int
	snapshots []SnapshotEntry
}

// NewHistory creates a History that retains up to maxSnapshots entries.
func NewHistory(maxSnapshots int) *History {
	if maxSnapshots <= 0 {
		maxSnapshots = 10
	}
	return &History{max: maxSnapshots}
}

// Record appends a new snapshot, evicting the oldest if at capacity.
func (h *History) Record(leases []vault.SecretLease) {
	h.mu.Lock()
	defer h.mu.Unlock()
	entry := SnapshotEntry{CapturedAt: time.Now(), Leases: leases}
	h.snapshots = append(h.snapshots, entry)
	if len(h.snapshots) > h.max {
		h.snapshots = h.snapshots[len(h.snapshots)-h.max:]
	}
}

// All returns a copy of all stored snapshots, oldest first.
func (h *History) All() []SnapshotEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]SnapshotEntry, len(h.snapshots))
	copy(out, h.snapshots)
	return out
}

// Latest returns the most recent snapshot, or false if empty.
func (h *History) Latest() (SnapshotEntry, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.snapshots) == 0 {
		return SnapshotEntry{}, false
	}
	return h.snapshots[len(h.snapshots)-1], true
}

// Len returns the number of stored snapshots.
func (h *History) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.snapshots)
}
