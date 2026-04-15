package vault

import (
	"encoding/json"
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// ClientV2 wraps the Vault API client with lease-focused helpers.
type ClientV2 struct {
	api *vaultapi.Client
}

// NewClientV2 creates a new Vault client from the given address and token.
func NewClientV2(address, token string) (*ClientV2, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	api, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}
	api.SetToken(token)
	return &ClientV2{api: api}, nil
}

// LookupLease fetches TTL information for a specific lease ID.
func (c *ClientV2) LookupLease(leaseID string) (*Lease, error) {
	raw, err := c.api.Logical().Write("sys/leases/lookup", map[string]interface{}{
		"lease_id": leaseID,
	})
	if err != nil {
		return nil, fmt.Errorf("looking up lease %q: %w", leaseID, err)
	}
	if raw == nil || raw.Data == nil {
		return nil, fmt.Errorf("empty response for lease %q", leaseID)
	}

	var ttlSec int64
	switch v := raw.Data["ttl"].(type) {
	case json.Number:
		ttlSec, _ = v.Int64()
	case float64:
		ttlSec = int64(v)
	}

	ttl := time.Duration(ttlSec) * time.Second
	return &Lease{
		ID:       leaseID,
		Path:     fmt.Sprint(raw.Data["path"]),
		TTL:      ttl,
		ExpireAt: time.Now().Add(ttl),
	}, nil
}

// ListLeases returns leases visible under the given sys/leases prefix.
func (c *ClientV2) ListLeases(prefix string) ([]Lease, error) {
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
		id := prefix + "/" + fmt.Sprint(k)
		lease, err := c.LookupLease(id)
		if err != nil {
			continue
		}
		leases = append(leases, *lease)
	}
	return leases, nil
}
