// Package filter provides lease filtering, sorting, and analysis utilities.
package filter

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// BaselineEntry records a lease's severity at a point in time.
type BaselineEntry struct {
	LeaseID  string    `json:"lease_id"`
	Path     string    `json:"path"`
	Severity string    `json:"severity"`
	CapturedAt time.Time `json:"captured_at"`
}

// BaselineDelta describes how a lease's severity has changed relative to baseline.
type BaselineDelta struct {
	Lease     vault.SecretLease
	Baseline  string // severity at baseline capture
	Current   string // current severity
	Changed   bool
	Worsened  bool   // true if current severity rank > baseline rank
}

// BaselineStore persists a severity baseline for later comparison.
type BaselineStore struct {
	mu      sync.RWMutex
	entries map[string]BaselineEntry // keyed by LeaseID
	path    string
}

// NewBaselineStore creates a BaselineStore backed by the given file path.
// If the file exists, existing entries are loaded automatically.
func NewBaselineStore(path string) (*BaselineStore, error) {
	bs := &BaselineStore{
		entries: make(map[string]BaselineEntry),
		path:    path,
	}
	if _, err := os.Stat(path); err == nil {
		if err := bs.load(); err != nil {
			return nil, fmt.Errorf("baseline: load %q: %w", path, err)
		}
	}
	return bs, nil
}

// Capture records the current severity of each lease as the new baseline,
// overwriting any previous entry for the same lease ID.
func (bs *BaselineStore) Capture(leases []vault.SecretLease) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	now := time.Now().UTC()
	for _, l := range leases {
		bs.entries[l.LeaseID] = BaselineEntry{
			LeaseID:    l.LeaseID,
			Path:       l.Path,
			Severity:   l.Severity,
			CapturedAt: now,
		}
	}
	return bs.persist()
}

// Compare returns a delta slice showing how each lease's severity compares
// to the stored baseline. Leases with no baseline entry are skipped.
func (bs *BaselineStore) Compare(leases []vault.SecretLease) []BaselineDelta {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	var deltas []BaselineDelta
	for _, l := range leases {
		entry, ok := bs.entries[l.LeaseID]
		if !ok {
			continue
		}
		changed := entry.Severity != l.Severity
		worsened := changed && severityRankBaseline(l.Severity) > severityRankBaseline(entry.Severity)
		deltas = append(deltas, BaselineDelta{
			Lease:    l,
			Baseline: entry.Severity,
			Current:  l.Severity,
			Changed:  changed,
			Worsened: worsened,
		})
	}
	sort.Slice(deltas, func(i, j int) bool {
		return deltas[i].Lease.LeaseID < deltas[j].Lease.LeaseID
	})
	return deltas
}

// Clear removes all baseline entries and persists the empty state.
func (bs *BaselineStore) Clear() error {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.entries = make(map[string]BaselineEntry)
	return bs.persist()
}

// persist writes the current entries to disk as JSON lines.
func (bs *BaselineStore) persist() error {
	if bs.path == "" {
		return nil
	}
	f, err := os.Create(bs.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(bs.entries)
}

// load reads entries from disk.
func (bs *BaselineStore) load() error {
	f, err := os.Open(bs.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&bs.entries)
}

// severityRankBaseline returns a numeric rank for comparison (higher = more severe).
func severityRankBaseline(s string) int {
	switch s {
	case "critical":
		return 3
	case "warn":
		return 2
	case "ok":
		return 1
	default:
		return 0
	}
}
