package filter_test

import (
	"strings"
	"testing"
	"time"

	"github.com/vaultpulse/internal/filter"
	"github.com/vaultpulse/internal/vault"
)

func makeTemplateLease(id, path, severity string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:  id,
		Path:     path,
		Severity: severity,
		TTL:      "2h",
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
}

func TestRenderTemplate_Default(t *testing.T) {
	leases := []vault.SecretLease{
		makeTemplateLease("id1", "secret/db", "critical"),
	}
	var sb strings.Builder
	if err := filter.RenderTemplate(leases, "", &sb); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "id1") {
		t.Errorf("expected lease id in output, got: %s", out)
	}
	if !strings.Contains(out, "secret/db") {
		t.Errorf("expected path in output, got: %s", out)
	}
}

func TestRenderTemplate_Custom(t *testing.T) {
	leases := []vault.SecretLease{
		makeTemplateLease("id2", "secret/api", "warn"),
	}
	tmpl := `{{range .}}LEASE:{{.LeaseID}}|SEV:{{.Severity}}\n{{end}}`
	var sb strings.Builder
	if err := filter.RenderTemplate(leases, tmpl, &sb); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "LEASE:id2") {
		t.Errorf("expected custom format, got: %s", out)
	}
	if !strings.Contains(out, "SEV:warn") {
		t.Errorf("expected severity in output, got: %s", out)
	}
}

func TestRenderTemplate_InvalidTemplate(t *testing.T) {
	leases := []vault.SecretLease{makeTemplateLease("id3", "p", "ok")}
	var sb strings.Builder
	err := filter.RenderTemplate(leases, "{{.Broken", &sb)
	if err == nil {
		t.Error("expected error for invalid template")
	}
}

func TestRenderTemplate_Empty(t *testing.T) {
	var sb strings.Builder
	if err := filter.RenderTemplate(nil, "", &sb); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sb.Len() != 0 {
		t.Errorf("expected empty output for nil leases, got: %s", sb.String())
	}
}
