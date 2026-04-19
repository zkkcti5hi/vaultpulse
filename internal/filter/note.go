package filter

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vaultpulse/internal/vault"
)

// NoteStore holds user-defined notes keyed by lease ID.
type NoteStore struct {
	mu    sync.RWMutex
	notes map[string]string
}

// NewNoteStore returns an empty NoteStore.
func NewNoteStore() *NoteStore {
	return &NoteStore{notes: make(map[string]string)}
}

// Set attaches a note to the given lease ID.
func (s *NoteStore) Set(leaseID, note string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notes[leaseID] = strings.TrimSpace(note)
}

// Get returns the note for a lease ID and whether it exists.
func (s *NoteStore) Get(leaseID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.notes[leaseID]
	return n, ok
}

// Delete removes the note for a lease ID.
func (s *NoteStore) Delete(leaseID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.notes, leaseID)
}

// ApplyNotes injects stored notes into the Metadata.Notes field of each lease.
func (s *NoteStore) ApplyNotes(leases []vault.SecretLease) []vault.SecretLease {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]vault.SecretLease, len(leases))
	for i, l := range leases {
		if note, ok := s.notes[l.LeaseID]; ok {
			if l.Metadata == nil {
				l.Metadata = map[string]string{}
			}
			l.Metadata["note"] = note
		}
		out[i] = l
	}
	return out
}

// List returns all lease IDs that have notes, sorted alphabetically.
func (s *NoteStore) List() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.notes))
	for id := range s.notes {
		ids = append(ids, id)
	}
	return ids
}

// String renders a human-readable summary of all notes.
func (s *NoteStore) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var sb strings.Builder
	for id, note := range s.notes {
		sb.WriteString(fmt.Sprintf("%s: %s\n", id, note))
	}
	return sb.String()
}
