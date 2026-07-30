// node_move_row_table_test.go — pins that MoveDispatch's row-identity tables (node/edge
// hit-test resolution + the mover-side row lookups) are built ONCE at load, and that
// node row order is DETERMINISTIC — specifically, directory-sorted (alphabetical) by node
// id (loadTree sorts nodeDirs, see its doc comment), NOT the order nodes are declared in
// the fixture. (The monolithic pre-tree form used to preserve JSON array/"seed" order;
// the tree form has no such concept — every load walks nodes/ sorted.) This is the
// MoveDispatch-side analogue of Buffer/row_order_test.go: proof that the row tables this
// package now owns produce identical, reproducible row indices for a representative graph.
// There is no port-row table any more (docs/channels-not-ports.md — a port has no buffer
// row of its own).

package Wiring

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"testing"

	T "github.com/dtauraso/wirefold/Trace"
)

func TestMoveDispatchRowTablesAreAlphabeticalByNodeID(t *testing.T) {
	// Three nodes are declared z-node, a-node, m-node (in THAT order) — deliberately NOT
	// alphabetical — so this fixture would fail if row assignment ever reverted to
	// declaration order, or became nondeterministic: only the alphabetical-by-id rule
	// makes every assertion below pass. One edge, so both node and edge row order can be
	// pinned.
	const topo = `{
	  "nodes": [
	    {"id":"z-node","type":"AimedSrc","scenePolarR":0,"scenePolarTheta":0,"scenePolarPhi":0,"cascadeEdges":["a-node"],"cascadeKinds":{"a-node":"AimedSink"}},
	    {"id":"a-node","type":"AimedSink","scenePolarR":50,"scenePolarTheta":1.5707963267948966,"scenePolarPhi":0,"cascadeEdges":["z-node"],"cascadeKinds":{"z-node":"AimedSrc"}},
	    {"id":"m-node","type":"AimedSink","scenePolarR":50,"scenePolarTheta":1.5707963267948966,"scenePolarPhi":3.14159}
	  ],
	  "edges": [
	    {"label":"e0","kind":"data","source":"z-node","sourceHandle":"Out","target":"a-node","targetHandle":"In"}
	  ]
	}`
	root := writeSpecTree(t, t.TempDir(), topo)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tr := T.New()
	clk := wire.NewRealClock()
	_, _, md, _, err := LoadTopology(ctx, root, tr, clk)
	if err != nil {
		t.Fatalf("LoadTopology: %v", err)
	}

	// Node rows: directory-sorted id order — a-node, m-node, z-node.
	wantNodes := []string{"a-node", "m-node", "z-node"}
	for row, id := range wantNodes {
		got, ok := md.LookupNodeRow(row)
		if !ok || got != id {
			t.Fatalf("LookupNodeRow(%d)=(%q,%v) want (%q,true)", row, got, ok, id)
		}
		if r, ok := md.NodeRowFor(id); !ok || r != int32(row) {
			t.Fatalf("NodeRowFor(%q)=(%d,%v) want (%d,true)", id, r, ok, row)
		}
	}
	if _, ok := md.LookupNodeRow(len(wantNodes)); ok {
		t.Fatalf("LookupNodeRow(%d) out of range: want ok=false", len(wantNodes))
	}

	// Edge rows: one edge, row 0.
	if l, ok := md.LookupEdgeRow(0); !ok || l != "e0" {
		t.Fatalf("LookupEdgeRow(0)=(%q,%v) want (e0,true)", l, ok)
	}
	if _, ok := md.LookupEdgeRow(1); ok {
		t.Fatalf("LookupEdgeRow(1) out of range: want ok=false")
	}
	if r, ok := md.EdgeRowForPair("a-node", "z-node"); !ok || r != 0 {
		t.Fatalf("EdgeRowForPair(a-node,z-node)=(%d,%v) want (0,true)", r, ok)
	}
	if r, ok := md.EdgeRowForPair("z-node", "a-node"); !ok || r != 0 {
		t.Fatalf("EdgeRowForPair(z-node,a-node)=(%d,%v) want (0,true) (order-independent)", r, ok)
	}
	if _, ok := md.EdgeRowForPair("a-node", "m-node"); ok {
		t.Fatalf("EdgeRowForPair(a-node,m-node): want ok=false (no such edge)")
	}
}
