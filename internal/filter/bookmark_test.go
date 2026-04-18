package filter_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeBookmarkLease(id, path string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Severity:  "ok",
	}
}

func TestBookmark_SaveAndGet(t *testing.T) {
	s := filter.NewBookmarkStore()
	leases := []vault.SecretLease{makeBookmarkLease("l1", "secret/a")}
	s.Save("snap1", leases)

	b, err := s.Get("snap1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Name != "snap1" {
		t.Errorf("expected name snap1, got %s", b.Name)
	}
	if len(b.Leases) != 1 {
		t.Errorf("expected 1 lease, got %d", len(b.Leases))
	}
}

func TestBookmark_GetNotFound(t *testing.T) {
	s := filter.NewBookmarkStore()
	_, err := s.Get("missing")
	if err == nil {
		t.Error("expected error for missing bookmark")
	}
}

func TestBookmark_Delete(t *testing.T) {
	s := filter.NewBookmarkStore()
	s.Save("snap1", nil)
	s.Delete("snap1")
	if s.Len() != 0 {
		t.Errorf("expected 0 bookmarks after delete, got %d", s.Len())
	}
}

func TestBookmark_List_Sorted(t *testing.T) {
	s := filter.NewBookmarkStore()
	s.Save("zebra", nil)
	s.Save("alpha", nil)
	s.Save("mango", nil)

	names := s.List()
	expected := []string{"alpha", "mango", "zebra"}
	for i, n := range names {
		if n != expected[i] {
			t.Errorf("index %d: expected %s got %s", i, expected[i], n)
		}
	}
}

func TestBookmark_Overwrite(t *testing.T) {
	s := filter.NewBookmarkStore()
	s.Save("snap", []vault.SecretLease{makeBookmarkLease("l1", "secret/a")})
	s.Save("snap", []vault.SecretLease{})

	b, _ := s.Get("snap")
	if len(b.Leases) != 0 {
		t.Errorf("expected 0 leases after overwrite, got %d", len(b.Leases))
	}
	if s.Len() != 1 {
		t.Errorf("expected 1 bookmark, got %d", s.Len())
	}
}
