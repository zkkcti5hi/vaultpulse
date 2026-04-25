package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// HeatmapCell represents a single cell in the expiry heatmap.
type HeatmapCell struct {
	Window    string
	Severity  string
	Count     int
	LeaseIDs  []string
}

// HeatmapOptions controls heatmap generation.
type HeatmapOptions struct {
	Windows []time.Duration // e.g. 1h, 6h, 24h, 72h
}

// DefaultHeatmapOptions returns sensible defaults.
func DefaultHeatmapOptions() HeatmapOptions {
	return HeatmapOptions{
		Windows: []time.Duration{
			1 * time.Hour,
			6 * time.Hour,
			24 * time.Hour,
			72 * time.Hour,
		},
	}
}

// Heatmap builds a grid of expiry counts bucketed by time window and severity.
func Heatmap(leases []vault.SecretLease, opts HeatmapOptions) []HeatmapCell {
	now := time.Now()
	var cells []HeatmapCell

	for _, w := range opts.Windows {
		counts := map[string]*HeatmapCell{}
		for _, l := range leases {
			ttl := l.ExpiresAt.Sub(now)
			if ttl <= 0 || ttl > w {
				continue
			}
			sev := l.Severity
			if sev == "" {
				sev = "ok"
			}
			key := sev
			if _, ok := counts[key]; !ok {
				counts[key] = &HeatmapCell{
					Window:   fmtDuration(w),
					Severity: sev,
				}
			}
			counts[key].Count++
			counts[key].LeaseIDs = append(counts[key].LeaseIDs, l.LeaseID)
		}
		for _, c := range counts {
			cells = append(cells, *c)
		}
	}

	sort.Slice(cells, func(i, j int) bool {
		if cells[i].Window != cells[j].Window {
			return cells[i].Window < cells[j].Window
		}
		return cells[i].Severity < cells[j].Severity
	})
	return cells
}

// PrintHeatmap writes an ASCII heatmap table to w.
func PrintHeatmap(cells []HeatmapCell, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, "%-10s %-10s %6s\n", "WINDOW", "SEVERITY", "COUNT")
	fmt.Fprintf(w, "%-10s %-10s %6s\n", "------", "--------", "-----")
	for _, c := range cells {
		fmt.Fprintf(w, "%-10s %-10s %6d\n", c.Window, c.Severity, c.Count)
	}
}

func fmtDuration(d time.Duration) string {
	h := int(d.Hours())
	if h < 24 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dd", h/24)
}
