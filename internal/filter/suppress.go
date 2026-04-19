package filter

import (
	"sync"
	"time"

	"github.com/vaultpulse/internal/vault"
)

// SuppressStore holds suppressed lease IDs with optional expiry.
type SuppressStore struct {
	mu      sync.RWMutex
	entries map[string]time.Time // zero Time means suppress forever
}

func NewSuppressStore() *SuppressStore {
	return &SuppressStore{entries: make(map[string]time.Time)}
}

// Suppress suppresses a lease ID until the given time. Pass zero time to suppress indefinitely.
func (s *SuppressStore) Suppress(leaseID string, until time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[leaseID] = until
}

// Unsuppress removes a lease from the suppress list.
func (s *SuppressStore) Unsuppress(leaseID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.entries[leaseID]; !ok {
		return false
	}
	delete(s.entries, leaseID)
	return true
}

// IsSuppressed reports whether a lease is currently suppressed.
func (s *SuppressStore) IsSuppressed(leaseID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	until, ok := s.entries[leaseID]
	if !ok {
		return false
	}
	if until.IsZero() {
		return true
	}
	return time.Now().Before(until)
}

// ApplySuppress filters out suppressed leases.
func (s *SuppressStore) ApplySuppress(leases []vault.SecretLease) []vault.SecretLease {
	out := make([]vault.SecretLease, 0, len(leases))
	for _, l := range leases {
		if !s.IsSuppressed(l.LeaseID) {
			out = append(out, l)
		}
	}
	return out
}

// List returns all currently active suppressions as a map of leaseID -> until.
func (s *SuppressStore) List() map[string]time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copy := make(map[string]time.Time, len(s.entries))
	for k, v := range s.entries {
		copy[k] = v
	}
	return copy
}
