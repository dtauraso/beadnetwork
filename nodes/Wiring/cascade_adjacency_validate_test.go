// cascade_adjacency_validate_test.go — validateCascadeEdges (topo_spec.go).
//
// Cascade adjacency is MANDATORY and must EQUAL the domain (edge) adjacency. That is what
// lets the drag fan (requantizeLocalPolars' neighborSetC) and the delta-forward fan
// (nodeMover.forwardDelta) read ONE stored list instead of two that can drift.

package Wiring

import (
	"strings"
	"testing"
)

// twoNodeSpec builds src--dst with one edge; cascade data is supplied by the caller so
// each case can break exactly one rule.
func twoNodeSpec(srcCascade []string, srcKinds map[string]string) topoSpec {
	return topoSpec{
		Nodes: []specNode{
			{ID: "src", Type: "SrcNode", CascadeEdges: srcCascade, CascadeKinds: srcKinds},
			{ID: "dst", Type: "SinkNode", CascadeEdges: []string{"src"},
				CascadeKinds: map[string]string{"src": "SrcNode"}},
		},
		Edges: []specEdge{{Label: "e0", Source: "src", Target: "dst"}},
	}
}

func TestValidateCascadeEdges(t *testing.T) {
	cases := []struct {
		name    string
		spec    topoSpec
		wantErr string // substring; "" means must pass
	}{
		{
			name: "equal to domain adjacency passes",
			spec: twoNodeSpec([]string{"dst"}, map[string]string{"dst": "SinkNode"}),
		},
		{
			name:    "missing file (nil cascade) is rejected",
			spec:    twoNodeSpec(nil, nil),
			wantErr: "is edge-adjacent to",
		},
		{
			name:    "self-loop is rejected",
			spec:    twoNodeSpec([]string{"src", "dst"}, map[string]string{"src": "SrcNode", "dst": "SinkNode"}),
			wantErr: "lists ITSELF",
		},
		{
			name:    "neighbor with no shared edge is rejected",
			spec:    twoNodeSpec([]string{"dst", "ghost"}, map[string]string{"dst": "SinkNode", "ghost": "SrcNode"}),
			wantErr: "unknown node",
		},
		{
			name:    "missing cascadeKinds entry is rejected",
			spec:    twoNodeSpec([]string{"dst"}, nil),
			wantErr: "no cascadeKinds entry",
		},
		{
			name:    "cascadeKinds disagreeing with the node's type is rejected",
			spec:    twoNodeSpec([]string{"dst"}, map[string]string{"dst": "WrongKind"}),
			wantErr: "but node",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateCascadeEdges(c.spec)
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("want pass, got error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("want error containing %q, got nil", c.wantErr)
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("want error containing %q, got: %v", c.wantErr, err)
			}
		})
	}
}

// TestValidateCascadeEdgesIsolatedNodeNeedsNone: a node with NO edges must carry NO
// cascade neighbors. This is why the rule is equality rather than "non-empty" — the
// non-empty draft made this case unexpressible and pushed fixtures into self-loops.
func TestValidateCascadeEdgesIsolatedNodeNeedsNone(t *testing.T) {
	spec := topoSpec{
		Nodes: []specNode{
			{ID: "src", Type: "SrcNode", CascadeEdges: []string{"dst"},
				CascadeKinds: map[string]string{"dst": "SinkNode"}},
			{ID: "dst", Type: "SinkNode", CascadeEdges: []string{"src"},
				CascadeKinds: map[string]string{"src": "SrcNode"}},
			{ID: "iso", Type: "SrcNode"}, // no edges, no cascade data
		},
		Edges: []specEdge{{Label: "e0", Source: "src", Target: "dst"}},
	}
	if err := validateCascadeEdges(spec); err != nil {
		t.Fatalf("isolated node with no cascade data must pass, got: %v", err)
	}
}
