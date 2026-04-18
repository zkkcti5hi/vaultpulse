package filter

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makePaginateLease(id string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      "secret/" + id,
		ExpiresAt: time.Now().Add(time.Hour),
		Severity:  "ok",
	}
}

func makeLeaseSlice(n int) []vault.SecretLease {
	leases := make([]vault.SecretLease, n)
	for i := 0; i < n; i++ {
		leases[i] = makePaginateLease(string(rune('a' + i)))
	}
	return leases
}

func TestPaginate_FirstPage(t *testing.T) {
	leases := makeLeaseSlice(25)
	p := Paginate(leases, 1, 10)
	if len(p.Items) != 10 {
		t.Fatalf("expected 10 items, got %d", len(p.Items))
	}
	if p.TotalPages != 3 {
		t.Fatalf("expected 3 total pages, got %d", p.TotalPages)
	}
	if !p.HasNext || p.HasPrev {
		t.Fatal("expected HasNext=true, HasPrev=false")
	}
}

func TestPaginate_LastPage(t *testing.T) {
	leases := makeLeaseSlice(25)
	p := Paginate(leases, 3, 10)
	if len(p.Items) != 5 {
		t.Fatalf("expected 5 items on last page, got %d", len(p.Items))
	}
	if p.HasNext || !p.HasPrev {
		t.Fatal("expected HasNext=false, HasPrev=true")
	}
}

func TestPaginate_Empty(t *testing.T) {
	p := Paginate([]vault.SecretLease{}, 1, 10)
	if len(p.Items) != 0 {
		t.Fatal("expected empty items")
	}
	if p.TotalPages != 1 {
		t.Fatalf("expected TotalPages=1, got %d", p.TotalPages)
	}
}

func TestPaginate_DefaultPageSize(t *testing.T) {
	leases := makeLeaseSlice(15)
	p := Paginate(leases, 1, 0)
	if p.PageSize != 10 {
		t.Fatalf("expected default page size 10, got %d", p.PageSize)
	}
}

func TestPaginate_OutOfRangePage(t *testing.T) {
	leases := makeLeaseSlice(5)
	p := Paginate(leases, 99, 10)
	if p.PageNumber != 1 {
		t.Fatalf("expected clamped page 1, got %d", p.PageNumber)
	}
}
