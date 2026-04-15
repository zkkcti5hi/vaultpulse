package vault

import (
	"sort"
	"time"
)

// ExpiryFilter holds criteria for filtering leases by expiry proximity.
type ExpiryFilter struct {
	// WarnBefore is the duration before expiry to consider a lease "expiring soon".
	WarnBefore time.Duration
	// CriticalBefore is the duration before expiry to consider a lease "critical".
	CriticalBefore time.Duration
}

// Severity indicates how urgent the lease expiry is.
type Severity string

const (
	SeverityOK       Severity = "ok"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// AnnotatedLease pairs a Lease with its computed severity.
type AnnotatedLease struct {
	Lease
	Severity Severity
}

// Annotate classifies each lease according to the provided ExpiryFilter.
func Annotate(leases []Lease, f ExpiryFilter) []AnnotatedLease {
	now := time.Now()
	result := make([]AnnotatedLease, 0, len(leases))

	for _, l := range leases {
		var sev Severity
		switch {
		case l.ExpireAt.Before(now.Add(f.CriticalBefore)):
			sev = SeverityCritical
		case l.ExpireAt.Before(now.Add(f.WarnBefore)):
			sev = SeverityWarning
		default:
			sev = SeverityOK
		}
		result = append(result, AnnotatedLease{Lease: l, Severity: sev})
	}

	// Sort: critical first, then warning, then ok; within group sort by expiry.
	sort.Slice(result, func(i, j int) bool {
		si, sj := severityRank(result[i].Severity), severityRank(result[j].Severity)
		if si != sj {
			return si > sj
		}
		return result[i].ExpireAt.Before(result[j].ExpireAt)
	})

	return result
}

func severityRank(s Severity) int {
	switch s {
	case SeverityCritical:
		return 2
	case SeverityWarning:
		return 1
	default:
		return 0
	}
}
