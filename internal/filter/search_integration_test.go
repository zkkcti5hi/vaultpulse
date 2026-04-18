package filter_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeIntSearchLease(id, path string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:  id,
		Path:     path,
		Severity: "ok",
		Expiry:   time.Now().Add(2 * time.Hour),
	}
}

func TestSearch_ThenSort_Integration(t *testing.T) {
	leases := []vault.SecretLease{
		makeIntSearchLease("aws/creds/dev/111", "aws/creds/dev"),
		makeIntSearchLease("aws/creds/prod/222", "aws/creds/prod"),
		makeIntSearchLease("database/creds/333", "database/creds"),
	}

	results := filter.Search(leases, filter.SearchOptions{Query: "aws"})
	if len(results) != 2 {
		t.Fatalf("expected 2 aws leases, got %d", len(results))
	}

	sorted := filter.Sort(results, filter.SortOptions{By: "path", Order: "asc"})
	if sorted[0].Path != "aws/creds/dev" {
		t.Errorf("expected aws/creds/dev first, got %s", sorted[0].Path)
	}
}

func TestSearch_ThenPaginate_Integration(t *testing.T) {
	var leases []vault.SecretLease
	for i := 0; i < 10; i++ {
		leases = append(leases, makeIntSearchLease(
			fmt.Sprintf("secret/app/%d", i),
			fmt.Sprintf("secret/app/%d", i),
		))
	}

	results := filter.Search(leases, filter.SearchOptions{Query: "app"})
	page, meta := filter.Paginate(results, filter.PaginateOptions{Page: 1, PageSize: 3})
	if len(page) != 3 {
		t.Fatalf("expected 3 per page, got %d", len(page))
	}
	if meta.Total != 10 {
		t.Fatalf("expected total 10, got %d", meta.Total)
	}
}
