// Package filter provides utilities for filtering, sorting, and scheduling
// Vault secret lease alerts.
//
// The schedule sub-feature (BuildSchedule, FilterScheduleByMinSeverity) allows
// callers to determine which leases fall within an upcoming notification window
// and should trigger alerts, optionally filtered by minimum severity level.
package filter
