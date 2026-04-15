package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "vaultpulse-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	raw := `
vault:
  address: "http://127.0.0.1:8200"
  token: "root"
monitor:
  interval: 10m
  secret_paths:
    - secret/myapp
alerts:
  warn_before: 48h
  critical_before: 12h
`
	path := writeTempConfig(t, raw)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Address != "http://127.0.0.1:8200" {
		t.Errorf("vault address = %q, want %q", cfg.Vault.Address, "http://127.0.0.1:8200")
	}
	if cfg.Alerts.WarnBefore != 48*time.Hour {
		t.Errorf("warn_before = %v, want 48h", cfg.Alerts.WarnBefore)
	}
	if cfg.Monitor.Interval != 10*time.Minute {
		t.Errorf("interval = %v, want 10m", cfg.Monitor.Interval)
	}
}

func TestLoad_Defaults(t *testing.T) {
	raw := `
vault:
  address: "http://127.0.0.1:8200"
  token: "root"
`
	path := writeTempConfig(t, raw)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Monitor.Interval != 5*time.Minute {
		t.Errorf("default interval = %v, want 5m", cfg.Monitor.Interval)
	}
	if cfg.Alerts.WarnBefore != 72*time.Hour {
		t.Errorf("default warn_before = %v, want 72h", cfg.Alerts.WarnBefore)
	}
	if cfg.Alerts.CriticalBefore != 24*time.Hour {
		t.Errorf("default critical_before = %v, want 24h", cfg.Alerts.CriticalBefore)
	}
}

func TestLoad_MissingAddress(t *testing.T) {
	raw := `
vault:
  token: "root"
`
	path := writeTempConfig(t, raw)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for missing vault.address, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
