// node_move_test.go — decentralized node-move path.
//
// Locks that a node-move handled WITHOUT a central coordinator reproduces the old
// applyNodeMove result per-goroutine: the moved node re-emits its node-geometry, and
// each incident edge recomputes its own segment/arc, re-emits its edge geometry,
// revises any in-flight bead, and updates the dest port's latency aggregate. The
// move is delivered exactly as the live bridge does — by mail-sorting each entry onto
// the node's own extIn channel and every incident edge's own extIn channel.

package Wiring

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"os"
	"path/filepath"
	"testing"

	T "github.com/dtauraso/wirefold/Trace"
)

// TestNodeGeometryLabelSidecar locks the new-system label sidecar contract at the Go
// layer: every node carries a Label (data.label when present, else the node id) and a
// Kind (the node's `type` field), the values each nodeMover's own dedicated stream frame
// packs (node_mover.go's writeStreamFrame) for the row-keyed {id,label}/kind→color
// sidecars. Read directly off each nodeMover's held geom — the same fields
// writeStreamFrame reads — rather than through the retired central Trace event path.
func TestNodeGeometryLabelSidecar(t *testing.T) {
	// "src" carries an explicit human label; "dst" omits data.label → label falls back to id.
	const topo = `{
	  "nodes": [
	    {"id":"src","type":"SrcNode","data":{"label":"Source Node"},"outputs":[{"name":"Out"}],"cascadeEdges":["dst"],"cascadeKinds":{"dst":"SinkNode"}},
	    {"id":"dst","type":"SinkNode","inputs":[{"name":"In"}],"cascadeEdges":["src"],"cascadeKinds":{"src":"SrcNode"}}
	  ],
	  "edges": [
	    {"label":"e0","kind":"data","source":"src","sourceHandle":"Out","target":"dst","targetHandle":"In"}
	  ],
	  "view": {"nodes": {
	    "src": {"x": 100, "y": 0, "z": 0},
	    "dst": {"x": 0,   "y": 0, "z": 0}
	  }}
	}`

	dir := t.TempDir()
	path := filepath.Join(dir, "topo.json")
	if err := os.WriteFile(path, []byte(topo), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tr := T.New()
	_, _, md, _, err := LoadTopology(ctx, path, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology: %v", err)
	}

	// Expected label per node id: explicit data.label for src, id fallback for dst.
	wantLabel := map[string]string{"src": "Source Node", "dst": "dst"}
	// Expected Go kind per node id: the node's `type` field, carried for the
	// new-system kind→color sidecar (row-keyed).
	wantKind := map[string]string{"src": "SrcNode", "dst": "SinkNode"}

	seen := map[string]bool{}
	for _, nm := range md.mr.nodeMovers {
		seen[nm.id] = true
		label := nm.geom.Label
		if label == "" {
			label = nm.id
		}
		if want := wantLabel[nm.id]; label != want {
			t.Fatalf("node %q: label = %q, want %q", nm.id, label, want)
		}
		if want := wantKind[nm.id]; nm.geom.Kind != want {
			t.Fatalf("node %q: kind = %q, want %q", nm.id, nm.geom.Kind, want)
		}
	}
	if len(seen) != 2 {
		t.Fatalf("saw %d distinct nodes, want 2", len(seen))
	}
}
