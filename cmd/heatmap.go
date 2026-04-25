package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

var heatmapWindows []string

func init() {
	heatmapCmd := &cobra.Command{
		Use:   "heatmap",
		Short: "Display a heatmap of lease expiry counts by time window and severity",
		RunE:  runHeatmap,
	}
	heatmapCmd.Flags().StringSliceVar(&heatmapWindows, "windows", []string{"1h", "6h", "24h", "72h"}, "Time windows for heatmap buckets")
	rootCmd.AddCommand(heatmapCmd)
}

func runHeatmap(cmd *cobra.Command, args []string) error {
	leases := staticHeatmapLeases()

	var windows []time.Duration
	for _, w := range heatmapWindows {
		d, err := time.ParseDuration(w)
		if err != nil {
			return fmt.Errorf("invalid window %q: %w", w, err)
		}
		windows = append(windows, d)
	}

	opts := filter.HeatmapOptions{Windows: windows}
	cells := filter.Heatmap(leases, opts)
	filter.PrintHeatmap(cells, os.Stdout)
	return nil
}

// staticHeatmapLeases returns example leases for demo/testing purposes.
func staticHeatmapLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "lease/1", Path: "secret/db", Severity: "critical", ExpiresAt: now.Add(30 * time.Minute)},
		{LeaseID: "lease/2", Path: "secret/api", Severity: "warn", ExpiresAt: now.Add(5 * time.Hour)},
		{LeaseID: "lease/3", Path: "secret/cache", Severity: "ok", ExpiresAt: now.Add(48 * time.Hour)},
		{LeaseID: "lease/4", Path: "secret/svc", Severity: "critical", ExpiresAt: now.Add(50 * time.Minute)},
	}
}
