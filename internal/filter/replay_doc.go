// Package filter provides replay functionality for vaultpulse.
//
// The ReplayStore records timestamped snapshots of lease states, allowing
// operators to review how the lease population changed over time. Events can
// be queried by nearest timestamp and printed as a human-readable timeline.
package filter
