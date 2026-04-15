package vault

import (
	"context"
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// SecretLease represents a Vault secret with lease metadata.
type SecretLease struct {
	LeaseID   string
	Path      string
	ExpiresAt time.Time
	Renewable bool
	TTL       time.Duration
}

// SecretFetcher defines the interface for fetching secret leases from Vault.
type SecretFetcher interface {
	ListLeases(ctx context.Context) ([]SecretLease, error)
}

// VaultSecretFetcher fetches active leases from a Vault client.
type VaultSecretFetcher struct {
	client *vaultapi.Client
}

// NewSecretFetcher creates a new VaultSecretFetcher.
func NewSecretFetcher(client *vaultapi.Client) *VaultSecretFetcher {
	return &VaultSecretFetcher{client: client}
}

// ListLeases queries the Vault sys/leases endpoint and returns active secret leases.
func (f *VaultSecretFetcher) ListLeases(ctx context.Context) ([]SecretLease, error) {
	secret, err := f.client.Logical().ListWithContext(ctx, "sys/leases/lookup")
	if err != nil {
		return nil, fmt.Errorf("listing leases: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return []SecretLease{}, nil
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []SecretLease{}, nil
	}

	var leases []SecretLease
	now := time.Now()

	for _, k := range keys {
		leaseID, ok := k.(string)
		if !ok {
			continue
		}

		info, err := f.lookupLease(ctx, leaseID)
		if err != nil {
			continue
		}

		leases = append(leases, SecretLease{
			LeaseID:   leaseID,
			Path:      info.path,
			ExpiresAt: now.Add(info.ttl),
			Renewable: info.renewable,
			TTL:       info.ttl,
		})
	}

	return leases, nil
}

type leaseInfo struct {
	path      string
	ttl       time.Duration
	renewable bool
}

func (f *VaultSecretFetcher) lookupLease(ctx context.Context, leaseID string) (leaseInfo, error) {
	secret, err := f.client.Logical().WriteWithContext(ctx, "sys/leases/lookup", map[string]interface{}{
		"lease_id": leaseID,
	})
	if err != nil {
		return leaseInfo{}, fmt.Errorf("looking up lease %s: %w", leaseID, err)
	}
	if secret == nil || secret.Data == nil {
		return leaseInfo{}, fmt.Errorf("empty response for lease %s", leaseID)
	}

	ttlRaw, _ := secret.Data["ttl"].(json.Number)
	ttlSecs, _ := ttlRaw.Int64()

	path, _ := secret.Data["path"].(string)
	renewable, _ := secret.Data["renewable"].(bool)

	return leaseInfo{
		path:      path,
		ttl:       time.Duration(ttlSecs) * time.Second,
		renewable: renewable,
	}, nil
}
