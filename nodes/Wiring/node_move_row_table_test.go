// node_move_row_table_test.go — pins that MoveDispatch's row-identity tables (node/edge
// hit-test resolution + the mover-side row lookups) are built ONCE at load, and that node
// row order is ROW ID = NODE ID - 1 — declared by the node's own id, never derived by
// sorting or declaration order. A gap in the id space leaves that row empty rather than
// shifting later rows down. This is the MoveDispatch-side analogue of
// Buffer/row_order_test.go: proof that the row tables this package now owns produce
// identical, reproducible row indices for a representative graph. There is no port-row
// table any more (docs/channels-not-ports.md — a port has no buffer row of its own).

package Wiring

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"testing"

	T "github.com/dtauraso/wirefold/Trace"
)

func TestMoveDispatchRowTablesUseNodeIDMinusOne(t *testing.T) {
	// Three nodes declared 30, 3, 15 — deliberately not adjacent — so this fixture would
	// fail if row assignment ever reverted to declaration order or to a dense sort-based
	// index, and only passes when each node's row is exactly its own id-1, with every other
	// row in [0, 30) left empty.
	const topo = `{
	  "nodes": [
	    {"id":"30","type":"AimedSrc","scenePolarR":0,"scenePolarTheta":0,"scenePolarPhi":0,"cascadeEdges":["3"],"cascadeKinds":{"3":"AimedSink"}},
	    {"id":"3","type":"AimedSink","scenePolarR":50,"scenePolarTheta":1.5707963267948966,"scenePolarPhi":0,"cascadeEdges":["30"],"cascadeKinds":{"30":"AimedSrc"}},
	    {"id":"15","type":"AimedSink","scenePolarR":50,"scenePolarTheta":1.5707963267948966,"scenePolarPhi":3.14159}
	  ],
	  "edges": [
	    {"label":"e0","kind":"data","source":"30","sourceHandle":"Out","target":"3","targetHandle":"In"}
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

	// Node rows: ROW ID = NODE ID - 1.
	wantRows := map[string]int{"3": 2, "15": 14, "30": 29}
	for id, row := range wantRows {
		got, ok := md.LookupNodeRow(row)
		if !ok || got != id {
			t.Fatalf("LookupNodeRow(%d)=(%q,%v) want (%q,true)", row, got, ok, id)
		}
		if r, ok := md.NodeRowFor(id); !ok || r != int32(row) {
			t.Fatalf("NodeRowFor(%q)=(%d,%v) want (%d,true)", id, r, ok, row)
		}
	}

	// A gap row (no node has id 1, so row 0 is empty) resolves to not-found, not to some
	// other node sliding down to fill it.
	if got, ok := md.LookupNodeRow(0); ok {
		t.Fatalf("LookupNodeRow(0) (gap row, no node id 1) = (%q,true), want ok=false", got)
	}

	// Out of the row space entirely (RowCount=30, rows 0..29) is also not-found.
	if _, ok := md.LookupNodeRow(30); ok {
		t.Fatalf("LookupNodeRow(30) out of range: want ok=false")
	}

	// Edge rows: one edge, row 0 (edge rows are dense spec order — not node-id-derived).
	if l, ok := md.LookupEdgeRow(0); !ok || l != "e0" {
		t.Fatalf("LookupEdgeRow(0)=(%q,%v) want (e0,true)", l, ok)
	}
	if _, ok := md.LookupEdgeRow(1); ok {
		t.Fatalf("LookupEdgeRow(1) out of range: want ok=false")
	}
	if r, ok := md.EdgeRowForPair("3", "30"); !ok || r != 0 {
		t.Fatalf("EdgeRowForPair(3,30)=(%d,%v) want (0,true)", r, ok)
	}
	if r, ok := md.EdgeRowForPair("30", "3"); !ok || r != 0 {
		t.Fatalf("EdgeRowForPair(30,3)=(%d,%v) want (0,true) (order-independent)", r, ok)
	}
	if _, ok := md.EdgeRowForPair("3", "15"); ok {
		t.Fatalf("EdgeRowForPair(3,15): want ok=false (no such edge)")
	}
}
