package filter

import (
	"fmt"
	"sort"
	"time"

	"github.com/your-org/vaultpulse/internal/vault"
)

// TrendPoint represents a severity count snapshot at a point in time.
type TrendPoint struct {
	At       time.Time
	Counts   map[string]int
}

// Trend analyses a slice of History snapshots and returns trend points.
func Trend(snapshots [][]vault.SecretLease) []TrendPoint {
	points := make([]TrendPoint, 0, len(snapshots))
	for i, leases := range snapshots {
		counts := map[string]int{"ok": 0, "warn": 0, "critical": 0}
		for _, l := range leases {
			sev := l.Severity
			if sev == "" {
				sev = "ok"
			}
			counts[sev]++
		}
		points = append(points, TrendPoint{
			At:     time.Now().Add(-time.Duration(len(snapshots)-i) * time.Minute),
			Counts: counts,
		})
	}
	return points
}

// PrintTrend writes a simple ASCII trend table to the provided writer.
func PrintTrend(points []TrendPoint) string {
	if len(points) == 0 {
		return "no trend data available\n"
	}
	sort.Slice(points, func(i, j int) bool {
		return points[i].At.Before(points[j].At)
	})
	out := fmt.Sprintf("%-25s %8s %8s %8s\n", "TIME", "OK", "WARN", "CRITICAL")
	out += fmt.Sprintf("%s\n", "--------------------------------------------------")
	for _, p := range points {
		out += fmt.Sprintf("%-25s %8d %8d %8d\n",
			p.At.Format(time.RFC3339),
			p.Counts["ok"],
			p.Counts["warn"],
			p.Counts["critical"],
		)
	}
	return out
}
