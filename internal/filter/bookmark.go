package filter

import (
	"fmt"
	"sort"
	"time"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// Bookmark saves a named snapshot of leases for later retrieval.
type Bookmark struct {
	Name      string
	SavedAt   time.Time
	Leases    []vault.SecretLease
}

// BookmarkStore holds named lease snapshots in memory.
type BookmarkStore struct {
	books map[string]Bookmark
}

// NewBookmarkStore returns an empty BookmarkStore.
func NewBookmarkStore() *BookmarkStore {
	return &BookmarkStore{books: make(map[string]Bookmark)}
}

// Save stores leases under the given name, overwriting any existing entry.
func (s *BookmarkStore) Save(name string, leases []vault.SecretLease) {
	s.books[name] = Bookmark{
		Name:    name,
		SavedAt: time.Now(),
		Leases:  leases,
	}
}

// Get retrieves a bookmark by name. Returns an error if not found.
func (s *BookmarkStore) Get(name string) (Bookmark, error) {
	b, ok := s.books[name]
	if !ok {
		return Bookmark{}, fmt.Errorf("bookmark %q not found", name)
	}
	return b, nil
}

// Delete removes a bookmark by name.
func (s *BookmarkStore) Delete(name string) {
	delete(s.books, name)
}

// List returns all bookmark names sorted alphabetically.
func (s *BookmarkStore) List() []string {
	names := make([]string, 0, len(s.books))
	for n := range s.books {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Len returns the number of stored bookmarks.
func (s *BookmarkStore) Len() int {
	return len(s.books)
}
