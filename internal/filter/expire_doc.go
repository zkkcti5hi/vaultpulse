// Package filter provides expire-window filtering utilities for vault leases.
// FilterByExpireWindow returns leases that will expire within a given duration.
// GroupByExpireWindow buckets leases into multiple named time windows.
// PrintExpireWindows renders a tabular summary to any io.Writer.
package filter
