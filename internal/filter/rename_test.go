package filter_test

import (
	"testing"
	"time"

	"github.com/nicholasgasior/vaultpulse/internal/filter"
	"github.com/nicholasgasior/vaultpulse/internal/vault"
)

func makeRenameLease(path, id string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:  id,
		Path:     path,
		TTL:      time.Hour,
		ExpireAt: time.Now().Add(time.Hour),
	}
}

func TestRename_NoAliases(t *testing.T) {
	leases := []vault.SecretLease{makeRenameLease("secret/db", "l1")}
	out := filter.Rename(leases, filter.RenameMap{})
	if out[0].Path != "secret/db" {
		t.Fatalf("expected path unchanged, got %s", out[0].Path)
	}
}

func TestRename_AppliesAlias(t *testing.T) {
	leases := []vault.SecretLease{
		makeRenameLease("secret/db", "l1"),
		makeRenameLease("secret/api", "l2"),
	}
	aliases := filter.RenameMap{"secret/db": "database"}
	out := filter.Rename(leases, aliases)
	if out[0].Path != "database" {
		t.Fatalf("expected 'database', got %s", out[0].Path)
	}
	if out[1].Path != "secret/api" {
		t.Fatalf("expected 'secret/api' unchanged, got %s", out[1].Path)
	}
}

func TestRename_OriginalUnmodified(t *testing.T) {
	original := []vault.SecretLease{makeRenameLease("secret/db", "l1")}
	aliases := filter.RenameMap{"secret/db": "database"}
	filter.Rename(original, aliases)
	if original[0].Path != "secret/db" {
		t.Fatal("original slice must not be mutated")
	}
}

func TestParseRenameFlag_Valid(t *testing.T) {
	m := filter.ParseRenameFlag([]string{"secret/db=database", "secret/api=api-key"})
	if m["secret/db"] != "database" {
		t.Fatalf("expected 'database', got %s", m["secret/db"])
	}
	if m["secret/api"] != "api-key" {
		t.Fatalf("expected 'api-key', got %s", m["secret/api"])
	}
}

func TestParseRenameFlag_SkipsInvalid(t *testing.T) {
	m := filter.ParseRenameFlag([]string{"nodivider", "secret/db=db"})
	if len(m) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(m))
	}
}

func TestParseRenameFlag_Empty(t *testing.T) {
	m := filter.ParseRenameFlag(nil)
	if len(m) != 0 {
		t.Fatal("expected empty map")
	}
}
