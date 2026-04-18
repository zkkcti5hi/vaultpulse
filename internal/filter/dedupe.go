package filter

import "github.com/yourusername/vaultpulse/internal/vault"

// Dedupe removes duplicate leases by LeaseID, keeping the first occurrence.
func Dedupe(leases []vault.SecretLease) []vault.SecretLease {
	if len(leases) == 0 {
		return leases
	}
	seen := make(map[string]struct{}, len(leases))
	out := make([]vault.SecretLease, 0, len(leases))
	for _, l := range leases {
		if _, ok := seen[l.LeaseID]; ok {
			continue
		}
		seen[l.LeaseID] = struct{}{}
		out = append(out, l)
	}
	return out
}
