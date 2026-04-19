package filter

import (
	"sort"
	"sync"

	"github.com/vaultpulse/internal/vault"
)

// PinStore holds leases that have been pinned by the user for quick access.
type PinStore struct {
	mu   sync.RWMutex
	pins map[string]vault.SecretLease
}

// NewPinStore creates an empty PinStore.
func NewPinStore() *PinStore {
	return &PinStore{pins: make(map[string]vault.SecretLease)}
}

// Pin adds or updates a lease in the store.
func (p *PinStore) Pin(lease vault.SecretLease) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pins[lease.LeaseID] = lease
}

// Unpin removes a lease by ID. Returns false if not found.
func (p *PinStore) Unpin(leaseID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.pins[leaseID]; !ok {
		return false
	}
	delete(p.pins, leaseID)
	return true
}

// Get returns a pinned lease by ID.
func (p *PinStore) Get(leaseID string) (vault.SecretLease, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	l, ok := p.pins[leaseID]
	return l, ok
}

// List returns all pinned leases sorted by LeaseID.
func (p *PinStore) List() []vault.SecretLease {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]vault.SecretLease, 0, len(p.pins))
	for _, l := range p.pins {
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].LeaseID < out[j].LeaseID
	})
	return out
}

// Len returns the number of pinned leases.
func (p *PinStore) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.pins)
}
