// node_move_test.go — decentralized node-move path.
//
// Locks that a node-move handled WITHOUT a central coordinator reproduces the old
// applyNodeMove result per-goroutine: the moved node re-emits its node-geometry, and
// each incident edge recomputes its own segment/arc, re-emits its edge geometry,
// revises any in-flight bead, and updates the dest port's latency aggregate. The
// move is delivered exactly as the live bridge does — by mail-sorting each entry onto
// the node's own extIn channel and every incident edge's own extIn channel.

package dispatch_test

import (
	"context"
	"strings"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/build"
	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

// TestNodeGeometryLabelSidecar locks the new-system label sidecar contract at the Go
// layer: every node carries a Label (data.label when present, else the node id) and a
// Kind (the node's `type` field), the values each nodeMover's own dedicated stream frame
// packs (node_mover.go's writeStreamFrame) for the row-keyed {id,label}/kind→color
// sidecars. Read directly off each nodeMover's held geom — the same fields
// writeStreamFrame reads — rather than through the retired central Trace event path.
func TestNodeGeometryLabelSidecar(t *testing.T) {
	// "1" carries an explicit human label; "2" omits data.label → label falls back to id.
	const topo = `{
	  "nodes": [
	    {"id":"1","type":"SrcNode","data":{"label":"Source Node"},"outputs":[{"name":"Out"}]},
	    {"id":"2","type":"SinkNode","inputs":[{"name":"In"}]}
	  ],
	  "edges": [
	    {"label":"e0","kind":"data","source":"1","sourceHandle":"Out","target":"2","targetHandle":"In"}
	  ],
	  "view": {"nodes": {
	    "1": {"x": 100, "y": 0, "z": 0},
	    "2": {"x": 0,   "y": 0, "z": 0}
	  }}
	}`

	root := dispatch.WriteSpecTree(t, t.TempDir(), topo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tr := T.New()
	_, _, md, _, err := build.LoadTopology(ctx, root, tr, clock.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology: %v", err)
	}

	// Expected label per node id: explicit data.label for "1", id fallback for "2".
	wantLabel := map[string]string{"1": "Source Node", "2": "2"}
	// Expected Go kind per node id: the node's `type` field, carried for the
	// new-system kind→color sidecar (row-keyed).
	wantKind := map[string]string{"1": "SrcNode", "2": "SinkNode"}

	seen := map[string]bool{}
	for _, nm := range md.MR.NodeGeoms() {
		id := nm.ID()
		seen[id] = true
		label := nm.Label()
		if label == "" {
			label = id
		}
		if want := wantLabel[id]; label != want {
			t.Fatalf("node %q: label = %q, want %q", id, label, want)
		}
		if want := wantKind[id]; nm.Kind() != want {
			t.Fatalf("node %q: kind = %q, want %q", id, nm.Kind(), want)
		}
	}
	if len(seen) != 2 {
		t.Fatalf("saw %d distinct nodes, want 2", len(seen))
	}
}

// TestNewMoveDispatchRejectsDanglingEdgeTarget locks the fix for the silent-zero-seed
// defect: an edge whose target (or source) node id has no entry in the geoms map — the
// shape a stale edge file left behind after its target node's directory was deleted by
// hand would produce, since in-edges are not indexed and nothing else catches it — must
// fail newMoveDispatch loudly, naming both the edge label and the missing node id,
// rather than silently seeding a degenerate 0,0,0->0,0,0 EdgeGeomSeed.
func TestNewMoveDispatchRejectsDanglingEdgeTarget(t *testing.T) {
	geoms := map[string]nodegeom.NodeGeom{
		"1": {},
	}
	edgeEndpoints := map[string]inputcodec.EdgeEndpoints{
		"e0": {Source: "1", Target: "9", SourceHandle: "Out", TargetHandle: "In"},
	}
	_, err := dispatch.NewMoveDispatch(geoms, edgeEndpoints, nil, nil, nil, clock.NewRealClock(), nil, 0)
	if err == nil {
		t.Fatal("NewMoveDispatch: want error for edge targeting a missing node, got nil")
	}
	if !strings.Contains(err.Error(), `"e0"`) {
		t.Fatalf("NewMoveDispatch error %q does not name the edge label \"e0\"", err.Error())
	}
	if !strings.Contains(err.Error(), `"9"`) {
		t.Fatalf("NewMoveDispatch error %q does not name the missing node id \"9\"", err.Error())
	}
}
