package filter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/yourusername/vaultpulse/internal/vault"
)

// TopologyNode represents a node in the secret path tree.
type TopologyNode struct {
	Name     string
	Path     string
	Leases   []vault.SecretLease
	Children map[string]*TopologyNode
}

// BuildTopology constructs a tree of TopologyNodes from a flat lease list.
func BuildTopology(leases []vault.SecretLease) *TopologyNode {
	root := &TopologyNode{Name: "/", Path: "/", Children: make(map[string]*TopologyNode)}
	for _, l := range leases {
		parts := strings.Split(strings.Trim(l.Path, "/"), "/")
		current := root
		built := ""
		for _, part := range parts {
			if part == "" {
				continue
			}
			built += "/" + part
			if _, ok := current.Children[part]; !ok {
				current.Children[part] = &TopologyNode{
					Name:     part,
					Path:     built,
					Children: make(map[string]*TopologyNode),
				}
			}
			current = current.Children[part]
		}
		current.Leases = append(current.Leases, l)
	}
	return root
}

// PrintTopology writes a tree-formatted topology to w (defaults to os.Stdout).
func PrintTopology(root *TopologyNode, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	printNode(w, root, "")
}

func printNode(w io.Writer, node *TopologyNode, indent string) {
	label := node.Name
	if len(node.Leases) > 0 {
		label += fmt.Sprintf(" [%d lease(s)]", len(node.Leases))
	}
	fmt.Fprintf(w, "%s%s\n", indent, label)

	keys := make([]string, 0, len(node.Children))
	for k := range node.Children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		child := node.Children[k]
		childIndent := indent + "  "
		if i < len(keys)-1 {
			fmt.Fprintf(w, "%s├─ ", indent)
		} else {
			fmt.Fprintf(w, "%s└─ ", indent)
		}
		printNode(w, child, childIndent)
	}
}
