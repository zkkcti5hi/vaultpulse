package filter

import (
	"strings"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// SearchOptions configures a text search over leases.
type SearchOptions struct {
	// Query is matched case-insensitively against lease ID and path.
	Query string
	// CaseSensitive disables case folding when true.
	CaseSensitive bool
}

// Search returns leases whose ID or path contains the query string.
// An empty query returns all leases unchanged.
func Search(leases []vault.SecretLease, opts SearchOptions) []vault.SecretLease {
	if opts.Query == "" {
		return leases
	}

	query := opts.Query
	if !opts.CaseSensitive {
		query = strings.ToLower(query)
	}

	var results []vault.SecretLease
	for _, l := range leases {
		id := l.LeaseID
		path := l.Path
		if !opts.CaseSensitive {
			id = strings.ToLower(id)
			path = strings.ToLower(path)
		}
		if strings.Contains(id, query) || strings.Contains(path, query) {
			results = append(results, l)
		}
	}
	return results
}
