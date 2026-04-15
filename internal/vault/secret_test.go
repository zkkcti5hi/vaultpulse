package vault

import (
	"context"
	"testing"
	"time"
)

// staticFetcher is a test double for SecretFetcher.
type staticFetcher struct {
	leases []SecretLease
	err    error
}

func (s *staticFetcher) ListLeases(_ context.Context) ([]SecretLease, error) {
	return s.leases, s.err
}

func makeSecretLease(leaseID, path string, ttl time.Duration, renewable bool) SecretLease {
	return SecretLease{
		LeaseID:   leaseID,
		Path:      path,
		ExpiresAt: time.Now().Add(ttl),
		Renewable: renewable,
		TTL:       ttl,
	}
}

func TestStaticFetcher_ReturnsList(t *testing.T) {
	expected := []SecretLease{
		makeSecretLease("lease/abc", "secret/data/myapp", 2*time.Hour, true),
		makeSecretLease("lease/xyz", "aws/creds/deploy", 30*time.Minute, false),
	}

	f := &staticFetcher{leases: expected}
	got, err := f.ListLeases(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(expected) {
		t.Fatalf("expected %d leases, got %d", len(expected), len(got))
	}
}

func TestStaticFetcher_ReturnsError(t *testing.T) {
	f := &staticFetcher{err: context.DeadlineExceeded}
	_, err := f.ListLeases(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSecretLease_ExpiresAt_IsAfterNow(t *testing.T) {
	lease := makeSecretLease("lease/test", "secret/data/test", 1*time.Hour, true)
	if !lease.ExpiresAt.After(time.Now()) {
		t.Errorf("ExpiresAt should be in the future, got %v", lease.ExpiresAt)
	}
}

func TestSecretLease_ZeroTTL(t *testing.T) {
	lease := makeSecretLease("lease/expired", "secret/data/old", 0, false)
	if lease.TTL != 0 {
		t.Errorf("expected TTL 0, got %v", lease.TTL)
	}
	// ExpiresAt should be approximately now
	diff := time.Since(lease.ExpiresAt)
	if diff > time.Second || diff < -time.Second {
		t.Errorf("ExpiresAt with zero TTL should be ~now, diff=%v", diff)
	}
}

func TestNewSecretFetcher_NotNil(t *testing.T) {
	// Verify constructor does not panic with nil client (struct wrapping only)
	f := NewSecretFetcher(nil)
	if f == nil {
		t.Fatal("expected non-nil fetcher")
	}
}
