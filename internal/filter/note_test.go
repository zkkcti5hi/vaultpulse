package filter

import (
	"testing"
	"time"

	"github.com/vaultpulse/internal/vault"
)

func makeNoteLease(id, path string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		ExpiresAt: time.Now().Add(time.Hour),
		Metadata:  map[string]string{},
	}
}

func TestNote_SetAndGet(t *testing.T) {
	s := NewNoteStore()
	s.Set("lease-1", "renew soon")
	note, ok := s.Get("lease-1")
	if !ok || note != "renew soon" {
		t.Fatalf("expected 'renew soon', got %q ok=%v", note, ok)
	}
}

func TestNote_GetNotFound(t *testing.T) {
	s := NewNoteStore()
	_, ok := s.Get("missing")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestNote_Delete(t *testing.T) {
	s := NewNoteStore()
	s.Set("lease-1", "note")
	s.Delete("lease-1")
	_, ok := s.Get("lease-1")
	if ok {
		t.Fatal("expected deleted")
	}
}

func TestNote_ApplyNotes_InjectsMetadata(t *testing.T) {
	s := NewNoteStore()
	s.Set("lease-1", "critical path")
	leases := []vault.SecretLease{
		makeNoteLease("lease-1", "secret/db"),
		makeNoteLease("lease-2", "secret/api"),
	}
	out := s.ApplyNotes(leases)
	if out[0].Metadata["note"] != "critical path" {
		t.Errorf("expected note injected, got %q", out[0].Metadata["note"])
	}
	if _, ok := out[1].Metadata["note"]; ok {
		t.Error("expected no note for lease-2")
	}
}

func TestNote_ApplyNotes_NilMetadata(t *testing.T) {
	s := NewNoteStore()
	s.Set("lease-x", "hello")
	l := vault.SecretLease{LeaseID: "lease-x", Path: "secret/x", ExpiresAt: time.Now().Add(time.Hour)}
	out := s.ApplyNotes([]vault.SecretLease{l})
	if out[0].Metadata["note"] != "hello" {
		t.Errorf("expected note on nil-metadata lease, got %q", out[0].Metadata["note"])
	}
}

func TestNote_List(t *testing.T) {
	s := NewNoteStore()
	s.Set("a", "note a")
	s.Set("b", "note b")
	ids := s.List()
	if len(ids) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(ids))
	}
}

func TestNote_TrimsWhitespace(t *testing.T) {
	s := NewNoteStore()
	s.Set("lease-1", "  spaced  ")
	n, _ := s.Get("lease-1")
	if n != "spaced" {
		t.Errorf("expected trimmed note, got %q", n)
	}
}
