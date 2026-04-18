package filter

import "github.com/your-org/vaultpulse/internal/vault"

// GroupBySeverity groups leases into a map keyed by severity label.
func GroupBySeverity(leases []vault.SecretLease) map[string][]vault.SecretLease {
	out := make(map[string][]vault.SecretLease)
	for _, l := range leases {
		out[l.Severity] = append(out[l.Severity], l)
	}
	return out
}

// GroupByPath groups leases by their top-level path prefix (first segment).
func GroupByPath(leases []vault.SecretLease) map[string][]vault.SecretLease {
	out := make(map[string][]vault.SecretLease)
	for _, l := range leases {
		prefix := pathPrefix(l.Path)
		out[prefix] = append(out[prefix], l)
	}
	return out
}

// pathPrefix returns the first path segment, e.g. "secret/foo/bar" -> "secret".
func pathPrefix(p string) string {
	for i, c := range p {
		if c == '/' && i > 0 {
			return p[:i]
		}
	}
	return p
}
