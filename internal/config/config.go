package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the full vaultpulse configuration.
type Config struct {
	Vault   VaultConfig   `yaml:"vault"`
	Alerts  AlertsConfig  `yaml:"alerts"`
	Monitor MonitorConfig `yaml:"monitor"`
}

// VaultConfig contains Vault connection settings.
type VaultConfig struct {
	Address   string `yaml:"address"`
	Token     string `yaml:"token"`
	Namespace string `yaml:"namespace"`
}

// AlertsConfig defines alert thresholds and destinations.
type AlertsConfig struct {
	// WarnBefore is the duration before expiry to start warning.
	WarnBefore time.Duration `yaml:"warn_before"`
	// CriticalBefore is the duration before expiry for critical alerts.
	CriticalBefore time.Duration `yaml:"critical_before"`
	SlackWebhook   string        `yaml:"slack_webhook"`
	EmailRecipient string        `yaml:"email_recipient"`
}

// MonitorConfig controls polling behaviour.
type MonitorConfig struct {
	Interval   time.Duration `yaml:"interval"`
	SecretPaths []string     `yaml:"secret_paths"`
}

// Load reads a YAML config file from the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	cfg := &Config{}
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(cfg); err != nil {
		return nil, fmt.Errorf("config: decode: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.Vault.Address == "" {
		return fmt.Errorf("config: vault.address is required")
	}
	if c.Vault.Token == "" {
		return fmt.Errorf("config: vault.token is required")
	}
	if c.Monitor.Interval <= 0 {
		c.Monitor.Interval = 5 * time.Minute
	}
	if c.Alerts.WarnBefore <= 0 {
		c.Alerts.WarnBefore = 72 * time.Hour
	}
	if c.Alerts.CriticalBefore <= 0 {
		c.Alerts.CriticalBefore = 24 * time.Hour
	}
	return nil
}
