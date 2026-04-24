package filter

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/your-org/vaultpulse/internal/vault"
)

// OutlierResult holds a lease flagged as an outlier along with its reason.
type OutlierResult struct {
	Lease  vault.SecretLease
	Reason string
}

// DetectOutliers identifies leases whose TTL deviates significantly from the
// mean TTL of the set. Leases with a TTL more than stddev multiplier (default 2)
// below the mean are flagged as outliers.
func DetectOutliers(leases []vault.SecretLease, multiplier float64) []OutlierResult {
	if len(leases) == 0 {
		return nil
	}
	if multiplier <= 0 {
		multiplier = 2.0
	}

	ttls := make([]float64, len(leases))
	var sum float64
	for i, l := range leases {
		v := l.TTL.Seconds()
		ttls[i] = v
		sum += v
	}
	mean := sum / float64(len(leases))

	var variance float64
	for _, v := range ttls {
		d := v - mean
		variance += d * d
	}
	variance /= float64(len(leases))
	stddev := 0.0
	if variance > 0 {
		stddev = variance
		// integer sqrt approximation via Newton's method
		stddev = sqrtFloat(variance)
	}

	threshold := mean - multiplier*stddev

	var results []OutlierResult
	for _, l := range leases {
		if l.TTL.Seconds() < threshold {
			reason := fmt.Sprintf("TTL %.0fs is %.1f stddevs below mean %.0fs",
				l.TTL.Seconds(), (mean-l.TTL.Seconds())/stddev, mean)
			results = append(results, OutlierResult{Lease: l, Reason: reason})
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Lease.TTL < results[j].Lease.TTL
	})
	return results
}

// PrintOutliers writes a formatted outlier report to w (defaults to os.Stdout).
func PrintOutliers(results []OutlierResult, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	if len(results) == 0 {
		fmt.Fprintln(w, "No outliers detected.")
		return
	}
	fmt.Fprintf(w, "%-40s %-12s %s\n", "PATH", "TTL", "REASON")
	fmt.Fprintf(w, "%-40s %-12s %s\n", "----", "---", "------")
	for _, r := range results {
		fmt.Fprintf(w, "%-40s %-12s %s\n", r.Lease.Path, r.Lease.TTL.String(), r.Reason)
	}
}

func sqrtFloat(x float64) float64 {
	if x == 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 20; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}
