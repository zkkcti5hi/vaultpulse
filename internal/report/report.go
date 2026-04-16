package report

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/user/vaultpulse/internal/vault"
)

// Format represents the output format for reports.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

// Reporter writes lease reports to an output stream.
type Reporter struct {
	out    io.Writer
	format Format
}

// NewReporter creates a Reporter writing to out in the given format.
func NewReporter(out io.Writer, format Format) *Reporter {
	if out == nil {
		out = os.Stdout
	}
	if format == "" {
		format = FormatTable
	}
	return &Reporter{out: out, format: format}
}

// Write renders leases to the configured output.
func (r *Reporter) Write(leases []vault.Lease) error {
	switch r.format {
	case FormatJSON:
		return r.writeJSON(leases)
	default:
		return r.writeTable(leases)
	}
}

func (r *Reporter) writeTable(leases []vault.Lease) error {
	w := tabwriter.NewWriter(r.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tSEVERITY\tEXPIRES IN\tLEASE ID")
	fmt.Fprintln(w, "----\t--------\t----------\t--------")
	for _, l := range leases {
		ttl := time.Until(l.ExpiresAt).Round(time.Second)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", l.Path, l.Severity, ttl, l.LeaseID)
	}
	return w.Flush()
}

func (r *Reporter) writeJSON(leases []vault.Lease) error {
	fmt.Fprintln(r.out, "[")
	for i, l := range leases {
		ttl := time.Until(l.ExpiresAt).Round(time.Second)
		comma := ","
		if i == len(leases)-1 {
			comma = ""
		}
		fmt.Fprintf(r.out, "  {\"path\":%q,\"severity\":%q,\"expires_in\":%q,\"lease_id\":%q}%s\n",
			l.Path, l.Severity, ttl.String(), l.LeaseID, comma)
	}
	fmt.Fprintln(r.out, "]")
	return nil
}
