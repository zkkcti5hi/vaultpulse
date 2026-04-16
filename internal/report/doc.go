// Package report provides formatted output for Vault lease data.
//
// It supports two formats:
//   - table: a human-readable tabular view (default)
//   - json:  a machine-readable JSON array
//
// Usage:
//
//	r := report.NewReporter(os.Stdout, report.FormatTable)
//	r.Write(leases)
package report
