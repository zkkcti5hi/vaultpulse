// Package filter provides utilities for filtering, sorting, deduplicating,
// grouping, summarising, paginating, and searching Vault secret leases.
//
// Search performs a case-insensitive (by default) substring match against
// lease IDs and paths, making it easy to locate specific leases in large
// result sets.
package filter
