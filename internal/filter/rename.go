package filter

import "github.com/nicholasgasior/vaultpulse/internal/vault"

// RenameMap maps old lease paths to new display aliases.
type RenameMap map[string]string

// Rename applies display aliases to leases based on a path→alias map.
// Leases whose path is not in the map are returned unchanged.
func Rename(leases []vault.SecretLease, aliases RenameMap) []vault.SecretLease {
	if len(aliases) == 0 {
		return leases
	}
	out := make([]vault.SecretLease, len(leases))
	for i, l := range leases {
		if alias, ok := aliases[l.Path]; ok {
			l.Path = alias
		}
		out[i] = l
	}
	return out
}

// ParseRenameFlag parses a slice of "old=new" strings into a RenameMap.
// Entries that do not contain "=" are silently skipped.
func ParseRenameFlag(pairs []string) RenameMap {
	m := make(RenameMap, len(pairs))
	for _, p := range pairs {
		for i := 0; i < len(p); i++ {
			if p[i] == '=' {
				oldPath := p[:i]
				newPath := p[i+1:]
				if oldPath != "" && newPath != "" {
					m[oldPath] = newPath
				}
				break
			}
		}
	}
	return m
}
