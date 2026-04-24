// Package filter provides lease filtering, sorting, grouping, and analysis
// utilities for vaultpulse.
//
// The forecast sub-feature predicts which leases will expire within a
// configurable look-ahead window, ordered by urgency. It respects the same
// severity ranking used elsewhere in the filter package so that operators can
// focus on warn/critical leases and ignore healthy ones.
package filter
