package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// ClusterResult holds leases grouped by a derived cluster key.
type ClusterResult struct {
	Key    string
	Leases []vault.SecretLease
}

// ClusterBy groups leases by a clustering strategy: "prefix", "severity", or "tag".
func ClusterBy(leases []vault.SecretLease, strategy string) []ClusterResult {
	index := make(map[string][]vault.SecretLease)

	for _, l := range leases {
		var key string
		switch strategy {
		case "severity":
			key = l.Severity
		case "tag":
			if len(l.Tags) == 0 {
				key = "untagged"
			} else {
				sorted := append([]string(nil), l.Tags...)
				sort.Strings(sorted)
				key = strings.Join(sorted, ",")
			}
		default: // "prefix"
			key = pathPrefix(l.Path)
		}
		index[key] = append(index[key], l)
	}

	results := make([]ClusterResult, 0, len(index))
	for k, v := range index {
		results = append(results, ClusterResult{Key: k, Leases: v})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Key < results[j].Key
	})
	return results
}

// PrintClusters writes a human-readable cluster summary to w.
func PrintClusters(clusters []ClusterResult, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	if len(clusters) == 0 {
		fmt.Fprintln(w, "no clusters")
		return
	}
	fmt.Fprintf(w, "%-30s  %s\n", "CLUSTER", "COUNT")
	fmt.Fprintln(w, strings.Repeat("-", 40))
	for _, c := range clusters {
		fmt.Fprintf(w, "%-30s  %d\n", c.Key, len(c.Leases))
	}
}
