package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

func init() {
	var baselineTTL time.Duration
	var tolerancePct float64
	var onlyExceeding bool

	cmd := &cobra.Command{
		Use:   "drift",
		Short: "Detect leases whose TTL deviates from a baseline",
		Long: `Compare each lease TTL against a configurable baseline and report
leases that fall outside the allowed tolerance window.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDrift(baselineTTL, tolerancePct, onlyExceeding)
		},
	}

	cmd.Flags().DurationVar(&baselineTTL, "baseline", 24*time.Hour, "expected lease TTL baseline (e.g. 24h)")
	cmd.Flags().Float64Var(&tolerancePct, "tolerance", 0.20, "allowed deviation from baseline as a fraction (0–1)")
	cmd.Flags().BoolVar(&onlyExceeding, "only-exceeding", false, "print only leases that exceed the tolerance")

	rootCmd.AddCommand(cmd)
}

func runDrift(baseline time.Duration, tolerance float64, onlyExceeding bool) error {
	leases := staticDriftLeases()
	if len(leases) == 0 {
		fmt.Fprintln(os.Stderr, "no leases to analyse")
		return nil
	}

	opts := filter.DriftOptions{
		BaselineTTL:  baseline,
		TolerancePct: tolerance,
		MinSeverity:  "ok",
	}

	results := filter.DetectDrift(leases, opts)

	if onlyExceeding {
		filtered := results[:0]
		for _, r := range results {
			if r.Exceeds {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	if len(results) == 0 {
		fmt.Println("no drift detected within current parameters")
		return nil
	}

	filter.PrintDrift(results, os.Stdout)
	return nil
}

// staticDriftLeases returns placeholder leases for demonstration.
// In production this would be replaced by a real Vault fetch.
func staticDriftLeases() []vault.SecretLease {
	now := time.Now()
	return []vault.SecretLease{
		{LeaseID: "lease/1", Path: "secret/db/prod", ExpiresAt: now.Add(25 * time.Hour), Severity: "ok"},
		{LeaseID: "lease/2", Path: "secret/db/staging", ExpiresAt: now.Add(1 * time.Hour), Severity: "critical"},
		{LeaseID: "lease/3", Path: "secret/api/key", ExpiresAt: now.Add(48 * time.Hour), Severity: "warn"},
	}
}
