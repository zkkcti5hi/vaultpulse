package filter_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeSnapLease(id, path string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
}

func tempSnapshotStore(t *testing.T) *filter.SnapshotStore {
	t.Helper()
	f := filepath.Join(t.TempDir(), "snapshots.json")
	return filter.NewSnapshotStore(f)
}

func TestSnapshot_SaveAndGet(t *testing.T) {
	store := tempSnapshotStore(t)
	leases := []vault.SecretLease{makeSnapLease("id1", "secret/a")}
	if err := store.Save("snap1", leases); err != nil {
		t.Fatalf("Save: %v", err)
	}
	snap, err := store.Get("snap1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if snap.Name != "snap1" || len(snap.Leases) != 1 {
		t.Errorf("unexpected snapshot: %+v", snap)
	}
}

func TestSnapshot_GetNotFound(t *testing.T) {
	store := tempSnapshotStore(t)
	_, err := store.Get("missing")
	if err == nil {
		t.Error("expected error for missing snapshot")
	}
}

func TestSnapshot_List(t *testing.T) {
	store := tempSnapshotStore(t)
	_ = store.Save("a", []vault.SecretLease{makeSnapLease("1", "p/a")})
	_ = store.Save("b", []vault.SecretLease{makeSnapLease("2", "p/b")})
	list := store.List()
	if len(list) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(list))
	}
}

func TestSnapshot_Delete(t *testing.T) {
	store := tempSnapshotStore(t)
	_ = store.Save("keep", []vault.SecretLease{makeSnapLease("1", "p/a")})
	_ = store.Save("remove", []vault.SecretLease{makeSnapLease("2", "p/b")})
	if err := store.Delete("remove"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	list := store.List()
	if len(list) != 1 || list[0].Name != "keep" {
		t.Errorf("unexpected list after delete: %+v", list)
	}
}

func TestSnapshot_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	store := filter.NewSnapshotStore(filepath.Join(dir, "new.json"))
	list := store.List()
	if len(list) != 0 {
		t.Errorf("expected empty list for new store")
	}
	_ = os.Remove(filepath.Join(dir, "new.json"))
}
