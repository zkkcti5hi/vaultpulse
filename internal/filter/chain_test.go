package filter_test

import (
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/filter"
	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeChainLease(id, path, severity string, ttl time.Duration) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   id,
		Path:      path,
		Severity:  severity,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func TestChain_EmptySteps(t *testing.T) {
	input := []vault.SecretLease{
		makeChainLease("a", "secret/a", "ok", time.Hour),
	}
	chain := filter.NewChain()
	out, trace := chain.Run(input)
	if len(out) != 1 {
		t.Fatalf("expected 1 lease, got %d", len(out))
	}
	if len(trace) != 0 {
		t.Fatalf("expected empty trace, got %d steps", len(trace))
	}
}

func TestChain_SingleStep(t *testing.T) {
	input := []vault.SecretLease{
		makeChainLease("a", "secret/a", "critical", time.Minute),
		makeChainLease("b", "secret/b", "ok", time.Hour),
	}
	chain := filter.NewChain().Add("filter-critical", func(ls []vault.SecretLease) []vault.SecretLease {
		var out []vault.SecretLease
		for _, l := range ls {
			if l.Severity == "critical" {
				out = append(out, l)
			}
		}
		return out
	})
	out, trace := chain.Run(input)
	if len(out) != 1 || out[0].LeaseID != "a" {
		t.Fatalf("expected lease 'a', got %+v", out)
	}
	if len(trace) != 1 || trace[0].Step != "filter-critical" {
		t.Fatalf("unexpected trace: %+v", trace)
	}
}

func TestChain_MultipleSteps_Composition(t *testing.T) {
	input := []vault.SecretLease{
		makeChainLease("a", "secret/prod/a", "critical", time.Minute),
		makeChainLease("b", "secret/dev/b", "critical", time.Minute),
		makeChainLease("c", "secret/prod/c", "ok", time.Hour),
	}
	chain := filter.NewChain().
		Add("only-critical", func(ls []vault.SecretLease) []vault.SecretLease {
			var out []vault.SecretLease
			for _, l := range ls {
				if l.Severity == "critical" {
					out = append(out, l)
				}
			}
			return out
		}).
		Add("only-prod", func(ls []vault.SecretLease) []vault.SecretLease {
			var out []vault.SecretLease
			for _, l := range ls {
				if strings.HasPrefix(l.Path, "secret/prod") {
					out = append(out, l)
				}
			}
			return out
		})
	out, trace := chain.Run(input)
	if len(out) != 1 || out[0].LeaseID != "a" {
		t.Fatalf("expected only lease 'a', got %+v", out)
	}
	if len(trace) != 2 {
		t.Fatalf("expected 2 trace entries, got %d", len(trace))
	}
	if trace[0].Step != "only-critical" || len(trace[0].Output) != 2 {
		t.Fatalf("unexpected first step trace: %+v", trace[0])
	}
}

func TestPrintTrace_ContainsStepNames(t *testing.T) {
	trace := []filter.ChainResult{
		{Step: "step-one", Output: []vault.SecretLease{}},
		{Step: "step-two", Output: []vault.SecretLease{{}, {}}},
	}
	out := filter.PrintTrace(trace)
	if !strings.Contains(out, "step-one") {
		t.Error("expected 'step-one' in trace output")
	}
	if !strings.Contains(out, "step-two") {
		t.Error("expected 'step-two' in trace output")
	}
	if !strings.Contains(out, "2") {
		t.Error("expected count '2' in trace output")
	}
}
