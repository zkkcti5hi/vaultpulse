package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpulse/internal/filter"
	"github.com/your-org/vaultpulse/internal/vault"
)

var (
	expireWindows []string
)

func init() {
	expireCmd := &cobra.Command{
		Use:   "expire",
		Short: "Show leases grouped by expiry time windows",
		RunE:  runExpire,
	}
	expireCmd.Flags().StringSliceVar(&expireWindows, "windows", []string{"1h", "6h", "24h"}, "Comma-separated list of time windows (e.g. 1h,6h,24h)")
	rootCmd.AddCommand(expireCmd)
}

func runExpire(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	client, err := vault.NewClientV2(cfg)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}
	fetcher := vault.NewSecretFetcher(client)
	leases, err := fetcher.Fetch(cmd.Context())
	if err != nil {
		return fmt.Errorf("fetch leases: %w", err)
	}

	windowMap := make(map[string]time.Duration, len(expireWindows))
	order := make([]string, 0, len(expireWindows))
	for _, w := range expireWindows {
		d, err := time.ParseDuration(w)
		if err != nil {
			return fmt.Errorf("invalid window %q: %w", w, err)
		}
		windowMap[w] = d
		order = append(order, w)
	}

	groups := filter.GroupByExpireWindow(leases, windowMap)
	filter.PrintExpireWindows(os.Stdout, groups, order)
	return nil
}
