package filter

import "github.com/yourusername/vaultpulse/internal/vault"

// Page holds a single page of leases and pagination metadata.
type Page struct {
	Items      []vault.SecretLease
	PageNumber int
	PageSize   int
	TotalItems int
	TotalPages int
	HasNext    bool
	HasPrev    bool
}

// Paginate splits leases into pages of the given size and returns the
// requested page (1-indexed). If pageSize <= 0 it defaults to 10.
// If page is out of range the nearest valid page is returned.
func Paginate(leases []vault.SecretLease, page, pageSize int) Page {
	if pageSize <= 0 {
		pageSize = 10
	}
	total := len(leases)
	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	return Page{
		Items:      leases[start:end],
		PageNumber: page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
