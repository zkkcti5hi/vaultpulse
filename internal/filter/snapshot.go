package filter

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// Snapshot holds a named point-in-time capture of leases.
type Snapshot struct {
	Name      string             `json:"name"`
	CreatedAt time.Time          `json:"created_at"`
	Leases    []vault.SecretLease `json:"leases"`
}

// SnapshotStore persists snapshots to a JSON file.
type SnapshotStore struct {
	path string
}

func NewSnapshotStore(path string) *SnapshotStore {
	return &SnapshotStore{path: path}
}

func (s *SnapshotStore) Save(name string, leases []vault.SecretLease) error {
	snap := Snapshot{Name: name, CreatedAt: time.Now(), Leases: leases}
	all, _ := s.loadAll()
	all = append(all, snap)
	return s.writeAll(all)
}

func (s *SnapshotStore) Get(name string) (Snapshot, error) {
	all, err := s.loadAll()
	if err != nil {
		return Snapshot{}, err
	}
	for _, snap := range all {
		if snap.Name == name {
			return snap, nil
		}
	}
	return Snapshot{}, fmt.Errorf("snapshot %q not found", name)
}

func (s *SnapshotStore) List() []Snapshot {
	all, _ := s.loadAll()
	return all
}

func (s *SnapshotStore) Delete(name string) error {
	all, err := s.loadAll()
	if err != nil {
		return err
	}
	filtered := all[:0]
	for _, snap := range all {
		if snap.Name != name {
			filtered = append(filtered, snap)
		}
	}
	return s.writeAll(filtered)
}

func (s *SnapshotStore) loadAll() ([]Snapshot, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var snaps []Snapshot
	return snaps, json.Unmarshal(data, &snaps)
}

func (s *SnapshotStore) writeAll(snaps []Snapshot) error {
	data, err := json.MarshalIndent(snaps, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}
