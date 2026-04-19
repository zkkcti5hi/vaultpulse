package filter

import (
	"bytes"
	"fmt"
	"io"
	"text/template"

	"github.com/vaultpulse/internal/vault"
)

// DefaultTemplate is used when no custom template is provided.
const DefaultTemplate = `{{range .}}[{{.Severity}}] {{.LeaseID}} ({{.Path}}) expires in {{.TTL}}
{{end}}`

// RenderTemplate renders leases using a Go text/template string.
func RenderTemplate(leases []vault.SecretLease, tmplStr string, w io.Writer) error {
	if tmplStr == "" {
		tmplStr = DefaultTemplate
	}
	tmpl, err := template.New("lease").Funcs(templateFuncs()).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, leases); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}
	_, err = w.Write(buf.Bytes())
	return err
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"upper": func(s string) string {
			if s == "" {
				return s
			}
			return fmt.Sprintf("%s", bytes.ToUpper([]byte(s)))
		},
		"default": func(def, val string) string {
			if val == "" {
				return def
			}
			return val
		},
	}
}
