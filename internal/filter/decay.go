package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/coolguy/vaultpulse/internal/vault"
)

// DecayEntry records how long a lease has been in a given severity.
type DecayEntry struct {
	LeaseID  string
	Path     string
	Severity string
	// Duration the lease has been at this severity level.
	Age time.Duration
	// Score representing urgency: higher means more stale / overdue attention.
	DecayScore float64
}

// Decay computes a decay score for each lease based on how long it has
// remained at its current severity without being resolved or suppressed.
// seenAt is a map of leaseID -> time the lease was first observed at its
// current severity. Leases not present in seenAt are treated as new (age=0).
func Decay(leases []vault.SecretLease, seenAt map[string]time.Time) []DecayEntry {
	now := time.Now()
	entries := make([]DecayEntry, 0, len(leases))

	for _, l := range leases {
		age := time.Duration(0)
		if t, ok := seenAt[l.LeaseID]; ok {
			age = now.Sub(t)
			if age < 0 {
				age = 0
			}
		}

		score := computeDecayScore(l.Severity, age)
		entries = append(entries, DecayEntry{
			LeaseID:    l.LeaseID,
			Path:       l.Path,
			Severity:   l.Severity,
			Age:        age,
			DecayScore: score,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].DecayScore > entries[j].DecayScore
	})

	return entries
}

// computeDecayScore returns a numeric score based on severity weight and age.
func computeDecayScore(severity string, age time.Duration) float64 {
	base := map[string]float64{
		"critical": 100.0,
		"warn":     50.0,
		"ok":       10.0,
	}[severity]

	hours := age.Hours()
	return base + hours*2.0
}

// PrintDecay writes a formatted decay table to w (defaults to os.Stdout).
func PrintDecay(entries []DecayEntry, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}

	fmt.Fprintf(w, "%-36s  %-30s  %-8s  %-12s  %s\n",
		"LEASE ID", "PATH", "SEVERITY", "AGE", "DECAY SCORE")
	fmt.Fprintf(w, "%s\n", "-------------------------------------------------------------------------------------------------------")

	for _, e := range entries {
		ageFmt := fmt.Sprintf("%.1fh", e.Age.Hours())
		fmt.Fprintf(w, "%-36s  %-30s  %-8s  %-12s  %.2f\n",
			e.LeaseID, e.Path, e.Severity, ageFmt, e.DecayScore)
	}
}
