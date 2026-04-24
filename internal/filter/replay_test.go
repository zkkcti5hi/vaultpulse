package filter

import (
	"bytes"
	"testing"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

func makeReplayLease(id, severity string) vault.SecretLease {
	return vault.SecretLease{
		LeaseID:  id,
		Path:     "secret/" + id,
		Severity: severity,
		TTL:      3600,
	}
}

func TestReplay_RecordAndLen(t *testing.T) {
	s := NewReplayStore()
	if s.Len() != 0 {
		t.Fatalf("expected 0, got %d", s.Len())
	}
	s.Record(time.Now(), []vault.SecretLease{makeReplayLease("a", "ok")})
	if s.Len() != 1 {
		t.Fatalf("expected 1, got %d", s.Len())
	}
}

func TestReplay_AtClosest(t *testing.T) {
	s := NewReplayStore()
	base := time.Now()
	s.Record(base.Add(-10*time.Minute), []vault.SecretLease{makeReplayLease("old", "ok")})
	s.Record(base.Add(-1*time.Minute), []vault.SecretLease{makeReplayLease("new", "critical")})

	e, ok := s.At(base)
	if !ok {
		t.Fatal("expected event")
	}
	if len(e.Leases) == 0 || e.Leases[0].LeaseID != "new" {
		t.Errorf("expected closest event, got %+v", e)
	}
}

func TestReplay_AtEmpty(t *testing.T) {
	s := NewReplayStore()
	_, ok := s.At(time.Now())
	if ok {
		t.Fatal("expected no event from empty store")
	}
}

func TestReplay_AllSorted(t *testing.T) {
	s := NewReplayStore()
	base := time.Now()
	s.Record(base.Add(5*time.Minute), []vault.SecretLease{makeReplayLease("c", "ok")})
	s.Record(base.Add(1*time.Minute), []vault.SecretLease{makeReplayLease("a", "ok")})
	s.Record(base.Add(3*time.Minute), []vault.SecretLease{makeReplayLease("b", "ok")})

	all := s.All()
	for i := 1; i < len(all); i++ {
		if all[i].At.Before(all[i-1].At) {
			t.Errorf("events not sorted at index %d", i)
		}
	}
}

func TestPrintReplay_ContainsHeaders(t *testing.T) {
	s := NewReplayStore()
	s.Record(time.Now(), []vault.SecretLease{makeReplayLease("x", "warn")})
	var buf bytes.Buffer
	PrintReplay(&buf, s.All())
	out := buf.String()
	for _, hdr := range []string{"TIME", "COUNT", "SEVERITY"} {
		if !bytes.Contains([]byte(out), []byte(hdr)) {
			t.Errorf("missing header %q in output: %s", hdr, out)
		}
	}
}

func TestPrintReplay_Empty(t *testing.T) {
	var buf bytes.Buffer
	PrintReplay(&buf, nil)
	if !bytes.Contains(buf.Bytes(), []byte("No replay")) {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}
