package alert

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/user/vaultpulse/internal/vault"
)

func makeLease(id, path, severity string, ttl time.Duration) vault.Lease {
	return vault.Lease{
		LeaseID:    id,
		Path:       path,
		ExpireTime: time.Now().Add(ttl),
		Severity:   severity,
	}
}

func captureStdout(fn func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestNotify_CriticalAlerted(t *testing.T) {
	n := NewNotifier(Config{Channel: ChannelStdout, MinLevel: "warning"})
	leases := []vault.Lease{
		makeLease("lease-1", "secret/db", "critical", 5*time.Minute),
	}
	out := captureStdout(func() {
		if err := n.Notify(leases); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "[critical]") {
		t.Errorf("expected critical alert in output, got: %s", out)
	}
	if !strings.Contains(out, "lease-1") {
		t.Errorf("expected lease ID in output, got: %s", out)
	}
}

func TestNotify_OkFiltered(t *testing.T) {
	n := NewNotifier(Config{Channel: ChannelStdout, MinLevel: "warning"})
	leases := []vault.Lease{
		makeLease("lease-ok", "secret/ok", "ok", 48*time.Hour),
	}
	out := captureStdout(func() {
		n.Notify(leases)
	})
	if strings.Contains(out, "lease-ok") {
		t.Errorf("expected ok lease to be filtered, got: %s", out)
	}
}

func TestNotify_Defaults(t *testing.T) {
	n := NewNotifier(Config{})
	if n.cfg.Channel != ChannelStdout {
		t.Errorf("expected default channel stdout, got %s", n.cfg.Channel)
	}
	if n.cfg.MinLevel != "warning" {
		t.Errorf("expected default minLevel warning, got %s", n.cfg.MinLevel)
	}
}

func TestNotify_EmptyLeases(t *testing.T) {
	n := NewNotifier(Config{Channel: ChannelStdout})
	out := captureStdout(func() {
		if err := n.Notify([]vault.Lease{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if out != "" {
		t.Errorf("expected no output for empty leases, got: %s", out)
	}
}

func TestNotify_WarningAlerted(t *testing.T) {
	n := NewNotifier(Config{Channel: ChannelStdout, MinLevel: "warning"})
	leases := []vault.Lease{
		makeLease("lease-warn", "secret/svc", "warning", 20*time.Minute),
	}
	out := captureStdout(func() {
		if err := n.Notify(leases); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "[warning]") {
		t.Errorf("expected warning alert in output, got: %s", out)
	}
	if !strings.Contains(out, "lease-warn") {
		t.Errorf("expected lease ID in output, got: %s", out)
	}
}
