// Package filter provides lease filtering, grouping, and analysis utilities.
//
// The heatmap module builds a two-dimensional grid of lease expiry counts
// bucketed by configurable time windows (e.g. 1h, 6h, 24h, 72h) and severity
// level. This gives operators a quick overview of upcoming expiry pressure
// across different urgency horizons.
package filter
