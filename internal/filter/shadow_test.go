package filter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeShadowLease(path, leaseID string, issued, expires time.Time) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:   leaseID,
		Path:      path,
		IssuedAt:  issued,
		ExpiresAt: expires,
	}
}

func TestDetectShadows_Empty(t *testing.T) {
	pairs := DetectShadows(nil, DefaultShadowOptions())
	if len(pairs) != 0 {
		t.Fatalf("expected 0 pairs, got %d", len(pairs))
	}
}

func TestDetectShadows_NoPairs_DifferentPaths(t *testing.T) {
	now := time.Now()
	leases := []vault.SecretLease{
		makeShadowLease("secret/a", "id-1", now.Add(-10*time.Minute), now.Add(50*time.Minute)),
		makeShadowLease("secret/b", "id-2", now.Add(-5*time.Minute), now.Add(55*time.Minute)),
	}
	pairs := DetectShadows(leases, DefaultShadowOptions())
	if len(pairs) != 0 {
		t.Fatalf("expected 0 pairs for different paths, got %d", len(pairs))
	}
}

func TestDetectShadows_DetectsPair(t *testing.T) {
	now := time.Now()
	// Older lease issued 20m ago, expires in 40m.
	// Newer lease issued 10m ago (overlaps older by 40m remaining).
	older := makeShadowLease("secret/db", "id-old", now.Add(-20*time.Minute), now.Add(40*time.Minute))
	newer := makeShadowLease("secret/db", "id-new", now.Add(-10*time.Minute), now.Add(50*time.Minute))

	leases := []vault.SecretLease{older, newer}
	pairs := DetectShadows(leases, DefaultShadowOptions())
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	if pairs[0].Older.LeaseID != "id-old" {
		t.Errorf("expected older lease id-old, got %s", pairs[0].Older.LeaseID)
	}
	if pairs[0].Newer.LeaseID != "id-new" {
		t.Errorf("expected newer lease id-new, got %s", pairs[0].Newer.LeaseID)
	}
	if pairs[0].Overlap <= 0 {
		t.Errorf("expected positive overlap, got %v", pairs[0].Overlap)
	}
}

func TestDetectShadows_BelowThreshold_Excluded(t *testing.T) {
	now := time.Now()
	// Overlap of only 1 minute, threshold is 5 minutes.
	older := makeShadowLease("secret/x", "id-a", now.Add(-30*time.Minute), now.Add(1*time.Minute))
	newer := makeShadowLease("secret/x", "id-b", now, now.Add(60*time.Minute))

	pairs := DetectShadows([]vault.SecretLease{older, newer}, DefaultShadowOptions())
	if len(pairs) != 0 {
		t.Fatalf("expected 0 pairs below threshold, got %d", len(pairs))
	}
}

func TestDetectShadows_SortedByOverlapDesc(t *testing.T) {
	now := time.Now()
	// Path A: large overlap.
	oa := makeShadowLease("secret/a", "a-old", now.Add(-60*time.Minute), now.Add(60*time.Minute))
	na := makeShadowLease("secret/a", "a-new", now.Add(-30*time.Minute), now.Add(90*time.Minute))
	// Path B: smaller overlap.
	ob := makeShadowLease("secret/b", "b-old", now.Add(-20*time.Minute), now.Add(10*time.Minute))
	nb := makeShadowLease("secret/b", "b-new", now.Add(-5*time.Minute), now.Add(55*time.Minute))

	pairs := DetectShadows([]vault.SecretLease{oa, na, ob, nb}, DefaultShadowOptions())
	if len(pairs) < 2 {
		t.Fatalf("expected at least 2 pairs, got %d", len(pairs))
	}
	if pairs[0].Overlap < pairs[1].Overlap {
		t.Errorf("pairs not sorted by overlap desc: %v < %v", pairs[0].Overlap, pairs[1].Overlap)
	}
}

func TestPrintShadows_ContainsHeaders(t *testing.T) {
	now := time.Now()
	older := makeShadowLease("secret/db", "id-old", now.Add(-20*time.Minute), now.Add(40*time.Minute))
	newer := makeShadowLease("secret/db", "id-new", now.Add(-10*time.Minute), now.Add(50*time.Minute))
	pairs := DetectShadows([]vault.SecretLease{older, newer}, DefaultShadowOptions())

	var buf bytes.Buffer
	opts := DefaultShadowOptions()
	opts.Out = &buf
	PrintShadows(pairs, opts)

	out := buf.String()
	for _, hdr := range []string{"PATH", "OLDER EXPIRES", "NEWER ISSUED", "OVERLAP"} {
		if !strings.Contains(out, hdr) {
			t.Errorf("output missing header %q", hdr)
		}
	}
}

func TestPrintShadows_Empty(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultShadowOptions()
	opts.Out = &buf
	PrintShadows(nil, opts)
	if !strings.Contains(buf.String(), "No shadow") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}
