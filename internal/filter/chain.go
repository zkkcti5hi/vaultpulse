package filter

import (
	"fmt"
	"strings"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// ChainStep represents a single named transformation step in a pipeline.
type ChainStep struct {
	Name string
	Fn   func([]vault.SecretLease) []vault.SecretLease
}

// Chain executes a sequence of filter/transform steps in order,
// returning the final slice and a summary of how many leases each step produced.
type Chain struct {
	steps []ChainStep
}

// ChainResult holds the output of each step for inspection.
type ChainResult struct {
	Step   string
	Output []vault.SecretLease
}

// NewChain creates an empty Chain.
func NewChain() *Chain {
	return &Chain{}
}

// Add appends a named step to the chain.
func (c *Chain) Add(name string, fn func([]vault.SecretLease) []vault.SecretLease) *Chain {
	c.steps = append(c.steps, ChainStep{Name: name, Fn: fn})
	return c
}

// Run executes all steps in order starting from input.
// It returns the final result and per-step trace.
func (c *Chain) Run(input []vault.SecretLease) ([]vault.SecretLease, []ChainResult) {
	current := input
	trace := make([]ChainResult, 0, len(c.steps))
	for _, step := range c.steps {
		current = step.Fn(current)
		trace = append(trace, ChainResult{Step: step.Name, Output: current})
	}
	return current, trace
}

// PrintTrace writes a human-readable summary of the chain execution to a string.
func PrintTrace(trace []ChainResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-24s  %s\n", "Step", "Count"))
	sb.WriteString(strings.Repeat("-", 36) + "\n")
	for _, r := range trace {
		sb.WriteString(fmt.Sprintf("%-24s  %d\n", r.Step, len(r.Output)))
	}
	return sb.String()
}
