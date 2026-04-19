package filter

import (
	"fmt"

	"github.com/your-org/vaultpulse/internal/vault"
)

// ThresholdConfig defines warning and critical TTL thresholds in hours.
type ThresholdConfig struct {
	WarnHours     float64
	CriticalHours float64
}

// DefaultThreshold provides sensible defaults.
var DefaultThreshold = ThresholdConfig{
	WarnHours:     72,
	CriticalHours: 24,
}

// ApplyThreshold re-annotates leases based on custom TTL thresholds,
// overriding the severity set during initial annotation.
func ApplyThreshold(leases []vault.SecretLease, cfg ThresholdConfig) []vault.SecretLease {
	result := make([]vault.SecretLease, len(leases))
	for i, l := range leases {
		hours := l.TTL.Hours()
		switch {
		case hours <= cfg.CriticalHours:
			l.Severity = "critical"
		case hours <= cfg.WarnHours:
			l.Severity = "warn"
		default:
			l.Severity = "ok"
		}
		result[i] = l
	}
	return result
}

// ParseThresholdFlag parses a "warn=72,critical=24" style string.
func ParseThresholdFlag(s string) (ThresholdConfig, error) {
	cfg := DefaultThreshold
	if s == "" {
		return cfg, nil
	}
	var warn, crit float64
	_, err := fmt.Sscanf(s, "warn=%f,critical=%f", &warn, &crit)
	if err != nil {
		return cfg, fmt.Errorf("invalid threshold format %q: expected warn=N,critical=N", s)
	}
	if warn <= 0 || crit <= 0 {
		return cfg, fmt.Errorf("threshold values must be positive")
	}
	if crit >= warn {
		return cfg, fmt.Errorf("critical threshold must be less than warn threshold")
	}
	cfg.WarnHours = warn
	cfg.CriticalHours = crit
	return cfg, nil
}
