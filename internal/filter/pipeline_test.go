package filter_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/vaultpulse/internal/filter"
	"github.com/vaultpulse/internal/vault"
)

func makePipelineLease(path, severity string, ttl time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:  "lease/" + path,
		Path:     path,
		Severity: severity,
		TTL:      ttl,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func TestPipeline_EmptySteps(t *testing.T) {
	leases := []vault.SecretLease{
		makePipelineLease("secret/a", "ok", time.Hour),
	}
	var buf bytes.Buffer
	p := filter.NewPipeline().WithWriter(&buf)
	out := p.Run(leases)
	if len(out) != 1 {
		t.Fatalf("expected 1 lease, got %d", len(out))
	}
	if buf.Len() != 0 {
		t.Errorf("expected no trace output for empty pipeline, got: %s", buf.String())
	}
}

func TestPipeline_SingleStep(t *testing.T) {
	leases := []vault.SecretLease{
		makePipelineLease("secret/a", "critical", time.Minute),
		makePipelineLease("secret/b", "ok", time.Hour),
	}
	var buf bytes.Buffer
	p := filter.NewPipeline().WithWriter(&buf)
	p.Add("filter-critical", func(ls []vault.SecretLease) []vault.SecretLease {
		var out []vault.SecretLease
		for _, l := range ls {
			if l.Severity == "critical" {
				out = append(out, l)
			}
		}
		return out
	})
	out := p.Run(leases)
	if len(out) != 1 {
		t.Fatalf("expected 1 lease, got %d", len(out))
	}
	if !strings.Contains(buf.String(), "filter-critical") {
		t.Errorf("expected trace to contain step name")
	}
}

func TestPipeline_MultipleSteps_Composition(t *testing.T) {
	leases := []vault.SecretLease{
		makePipelineLease("secret/a", "critical", time.Minute),
		makePipelineLease("secret/b", "warn", 30*time.Minute),
		makePipelineLease("secret/c", "ok", time.Hour),
	}
	p := filter.NewPipeline().WithWriter(&bytes.Buffer{})
	p.Add("drop-ok", func(ls []vault.SecretLease) []vault.SecretLease {
		var out []vault.SecretLease
		for _, l := range ls {
			if l.Severity != "ok" {
				out = append(out, l)
			}
		}
		return out
	})
	p.Add("drop-warn", func(ls []vault.SecretLease) []vault.SecretLease {
		var out []vault.SecretLease
		for _, l := range ls {
			if l.Severity != "warn" {
				out = append(out, l)
			}
		}
		return out
	})
	out := p.Run(leases)
	if len(out) != 1 || out[0].Path != "secret/a" {
		t.Errorf("expected only critical lease, got %+v", out)
	}
}

func TestPipeline_StepNames(t *testing.T) {
	p := filter.NewPipeline().WithWriter(&bytes.Buffer{})
	p.Add("step-one", func(ls []vault.SecretLease) []vault.SecretLease { return ls })
	p.Add("step-two", func(ls []vault.SecretLease) []vault.SecretLease { return ls })
	names := p.StepNames()
	if len(names) != 2 || names[0] != "step-one" || names[1] != "step-two" {
		t.Errorf("unexpected step names: %v", names)
	}
}

func TestParsePipelineSteps(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{"filter,sort,dedupe", []string{"filter", "sort", "dedupe"}},
		{" filter , sort ", []string{"filter", "sort"}},
		{"", nil},
	}
	for _, c := range cases {
		got := filter.ParsePipelineSteps(c.input)
		if len(got) != len(c.expected) {
			t.Errorf("input %q: expected %v, got %v", c.input, c.expected, got)
			continue
		}
		for i := range got {
			if got[i] != c.expected[i] {
				t.Errorf("input %q index %d: expected %q, got %q", c.input, i, c.expected[i], got[i])
			}
		}
	}
}

func TestPrintPipeline_ContainsStepNames(t *testing.T) {
	p := filter.NewPipeline().WithWriter(&bytes.Buffer{})
	p.Add("alpha", func(ls []vault.SecretLease) []vault.SecretLease { return ls })
	p.Add("beta", func(ls []vault.SecretLease) []vault.SecretLease { return ls })
	var buf bytes.Buffer
	filter.PrintPipeline(p, &buf)
	out := buf.String()
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Errorf("expected step names in output, got: %s", out)
	}
}
