package filter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	"github.com/your-org/vaultpulse/internal/vault"
)

// Format represents the export output format.
type Format string

const (
	FormatCSV  Format = "csv"
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Export writes leases to w in the requested format.
func Export(w io.Writer, leases []vault.SecretLease, format Format) error {
	switch format {
	case FormatCSV:
		return exportCSV(w, leases)
	case FormatJSON:
		return exportJSON(w, leases)
	default:
		return exportText(w, leases)
	}
}

func exportCSV(w io.Writer, leases []vault.SecretLease) error {
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"lease_id", "path", "severity", "expires_at", "ttl_seconds"})
	for _, l := range leases {
		_ = cw.Write([]string{
			l.LeaseID,
			l.Path,
			l.Severity,
			l.ExpiresAt.Format("2006-01-02T15:04:05Z"),
			fmt.Sprintf("%d", int(l.TTL.Seconds())),
		})
	}
	cw.Flush()
	return cw.Error()
}

func exportJSON(w io.Writer, leases []vault.SecretLease) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(leases)
}

func exportText(w io.Writer, leases []vault.SecretLease) error {
	for _, l := range leases {
		_, err := fmt.Fprintf(w, "[%s] %s (expires: %s, ttl: %s)\n",
			l.Severity, l.Path, l.ExpiresAt.Format("2006-01-02T15:04:05Z"), l.TTL)
		if err != nil {
			return err
		}
	}
	return nil
}
