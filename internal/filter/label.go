package filter

import "github.com/nicholasgasior/vaultpulse/internal/vault"

// LabelStore holds user-defined labels keyed by lease ID.
type LabelStore struct {
	labels map[string][]string
}

// NewLabelStore returns an empty LabelStore.
func NewLabelStore() *LabelStore {
	return &LabelStore{labels: make(map[string][]string)}
}

// Add attaches a label to a lease ID. Duplicate labels are ignored.
func (s *LabelStore) Add(leaseID, label string) {
	for _, l := range s.labels[leaseID] {
		if l == label {
			return
		}
	}
	s.labels[leaseID] = append(s.labels[leaseID], label)
}

// Remove detaches a label from a lease ID.
func (s *LabelStore) Remove(leaseID, label string) {
	current := s.labels[leaseID]
	updated := current[:0]
	for _, l := range current {
		if l != label {
			updated = append(updated, l)
		}
	}
	s.labels[leaseID] = updated
}

// Get returns all labels for a lease ID.
func (s *LabelStore) Get(leaseID string) []string {
	return s.labels[leaseID]
}

// FilterByLabel returns leases that have the given label attached.
func (s *LabelStore) FilterByLabel(leases []vault.SecretLease, label string) []vault.SecretLease {
	var out []vault.SecretLease
	for _, l := range leases {
		for _, lbl := range s.labels[l.LeaseID] {
			if lbl == label {
				out = append(out, l)
				break
			}
		}
	}
	return out
}
