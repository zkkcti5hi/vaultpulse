package vault

import (
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client with lease-focused helpers.
type Client struct {
	api *vaultapi.Client
}

// Lease represents a single secret lease returned by Vault.
type Lease struct {
	ID        string
	Path      string
	TTL       time.Duration
	ExpireAt  time.Time
}

// NewClient creates a new Vault client from the given address and token.
func NewClient(address, token string) (*Client, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	api, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}

	api.SetToken(token)

	return &Client{api: api}, nil
}

// ListLeases returns all renewable leases visible under the given prefix.
func (c *Client) ListLeases(prefix string) ([]Lease, error) {
	sys := c.api.Sys()

	resp, err := sys.RenewSelf(0)
	_ = resp
	if err != nil {
		// non-fatal: token may not be renewable
	}

	raw, err := c.api.Logical().List("sys/leases/lookup/" + prefix)
	if err != nil {
		return nil, fmt.Errorf("listing leases at %q: %w", prefix, err)
	}
	if raw == nil || raw.Data == nil {
		return nil, nil
	}

	keys, ok := raw.Data["keys"].([]interface{})
	if !ok {
		return nil, nil
	}

	var leases []Lease
	for _, k := range keys {
		id := prefix + fmt.Sprint(k)
		lease, err := c.LookupLease(id)
		if err != nil {
			continue
		}
		leases = append(leases, *lease)
	}
	return leases, nil
}

// LookupLease fetches TTL information for a specific lease ID.
func (c *Client) LookupLease(leaseID string) (*Lease, error) {
	raw, err := c.api.Logical().Write("sys/leases/lookup", map[string]interface{}{
		"lease_id": leaseID,
	})
	if err != nil {
		return nil, fmt.Errorf("looking up lease %q: %w", leaseID, err)
	}
	if raw == nil || raw.Data == nil {
		return nil, fmt.Errorf("empty response for lease %q", leaseID)
	}

	ttlRaw, _ := raw.Data["ttl"].(json.Number)
	ttlSec, _ := ttlRaw.Int64()
	ttl := time.Duration(ttlSec) * time.Second

	return &Lease{
		ID:       leaseID,
		Path:     fmt.Sprint(raw.Data["path"]),
		TTL:      ttl,
		ExpireAt: time.Now().Add(ttl),
	}, nil
}
