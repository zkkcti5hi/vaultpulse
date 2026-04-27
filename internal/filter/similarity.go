package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// SimilarityOptions controls how lease similarity is computed.
type SimilarityOptions struct {
	// MinScore is the minimum Jaccard-like score (0.0–1.0) to include a pair.
	MinScore float64
	// MaxResults caps the number of pairs returned.
	MaxResults int
}

// DefaultSimilarityOptions returns sensible defaults.
func DefaultSimilarityOptions() SimilarityOptions {
	return SimilarityOptions{
		MinScore:   0.5,
		MaxResults: 20,
	}
}

// SimilarPair represents two leases that share structural similarity.
type SimilarPair struct {
	A     vault.SecretLease
	B     vault.SecretLease
	Score float64
}

// FindSimilar computes pairwise similarity between leases based on path
// segments and tags, returning pairs whose score meets opts.MinScore.
func FindSimilar(leases []vault.SecretLease, opts SimilarityOptions) []SimilarPair {
	var pairs []SimilarPair
	for i := 0; i < len(leases); i++ {
		for j := i + 1; j < len(leases); j++ {
			s := jaccardScore(leases[i], leases[j])
			if s >= opts.MinScore {
				pairs = append(pairs, SimilarPair{A: leases[i], B: leases[j], Score: s})
			}
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Score > pairs[j].Score
	})
	if opts.MaxResults > 0 && len(pairs) > opts.MaxResults {
		pairs = pairs[:opts.MaxResults]
	}
	return pairs
}

// PrintSimilarity writes a human-readable similarity report to w.
func PrintSimilarity(pairs []SimilarPair, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	if len(pairs) == 0 {
		fmt.Fprintln(w, "no similar lease pairs found")
		return
	}
	fmt.Fprintf(w, "%-45s %-45s %s\n", "LEASE A", "LEASE B", "SCORE")
	fmt.Fprintln(w, strings.Repeat("-", 100))
	for _, p := range pairs {
		fmt.Fprintf(w, "%-45s %-45s %.2f\n", truncatePath(p.A.Path, 44), truncatePath(p.B.Path, 44), p.Score)
	}
}

func jaccardScore(a, b vault.SecretLease) float64 {
	setA := tokenSet(a)
	setB := tokenSet(b)
	intersection := 0
	for k := range setA {
		if setB[k] {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func tokenSet(l vault.SecretLease) map[string]bool {
	set := make(map[string]bool)
	for _, seg := range strings.Split(l.Path, "/") {
		if seg != "" {
			set[strings.ToLower(seg)] = true
		}
	}
	for _, tag := range l.Metadata["tags"] {
		set["tag:"+strings.ToLower(tag)] = true
	}
	return set
}

func truncatePath(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
