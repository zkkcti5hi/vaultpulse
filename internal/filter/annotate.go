package filter

import (
	"strings"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// AnnotateOptions controls which annotations are applied.
type AnnotateOptions struct {
	AddTags    []string
	AddLabels  []string
	NotePrefix string
}

// AnnotateResult holds the original lease plus applied metadata.
type AnnotateResult struct {
	Lease   vault.SecretLease
	Tags    []string
	Labels  []string
	Note    string
}

// AnnotateLeases applies tags, labels, and a note prefix to each lease.
// Tags and labels are deduplicated (case-insensitive).
func AnnotateLeases(leases []vault.SecretLease, opts AnnotateOptions) []AnnotateResult {
	results := make([]AnnotateResult, 0, len(leases))
	for _, l := range leases {
		ar := AnnotateResult{
			Lease:  l,
			Tags:   dedupStrings(opts.AddTags),
			Labels: dedupStrings(opts.AddLabels),
		}
		if opts.NotePrefix != "" {
			ar.Note = opts.NotePrefix + ": " + l.LeaseID
		}
		results = append(results, ar)
	}
	return results
}

// dedupStrings returns a new slice with duplicates removed (case-insensitive).
func dedupStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		key := strings.ToLower(s)
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}
