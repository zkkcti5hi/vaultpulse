package filter

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vaultpulse/internal/vault"
)

// PipelineStep is a named transformation applied to a lease slice.
type PipelineStep struct {
	Name string
	Fn   func([]vault.SecretLease) []vault.SecretLease
}

// Pipeline executes an ordered sequence of steps, threading the output of
// each step into the next. It optionally emits per-step counts to w.
type Pipeline struct {
	steps []PipelineStep
	w     io.Writer
}

// NewPipeline returns an empty Pipeline that writes trace output to os.Stdout.
func NewPipeline() *Pipeline {
	return &Pipeline{w: os.Stdout}
}

// WithWriter overrides the trace writer.
func (p *Pipeline) WithWriter(w io.Writer) *Pipeline {
	p.w = w
	return p
}

// Add appends a named step to the pipeline.
func (p *Pipeline) Add(name string, fn func([]vault.SecretLease) []vault.SecretLease) *Pipeline {
	p.steps = append(p.steps, PipelineStep{Name: name, Fn: fn})
	return p
}

// Run executes all steps in order and returns the final lease slice.
// Each step name and output count is written to the configured writer.
func (p *Pipeline) Run(leases []vault.SecretLease) []vault.SecretLease {
	current := leases
	for _, step := range p.steps {
		current = step.Fn(current)
		fmt.Fprintf(p.w, "[pipeline] %-24s → %d leases\n", step.Name, len(current))
	}
	return current
}

// StepNames returns the names of all registered steps.
func (p *Pipeline) StepNames() []string {
	names := make([]string, len(p.steps))
	for i, s := range p.steps {
		names[i] = s.Name
	}
	return names
}

// PrintPipeline writes a summary of the pipeline configuration to w.
func PrintPipeline(p *Pipeline, w io.Writer) {
	fmt.Fprintln(w, "Pipeline steps:")
	for i, name := range p.StepNames() {
		fmt.Fprintf(w, "  %2d. %s\n", i+1, name)
	}
	if len(p.steps) == 0 {
		fmt.Fprintln(w, "  (none)")
	}
}

// ParsePipelineSteps parses a comma-separated list of step names into a slice.
func ParsePipelineSteps(raw string) []string {
	var out []string
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
