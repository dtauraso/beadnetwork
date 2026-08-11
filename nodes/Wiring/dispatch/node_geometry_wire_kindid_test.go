// node_geometry_wire_kindid_test.go — closes the coverage hole recorded in
// docs/planning/movedispatch-decomposition.md §19 ("Deliberate breaks, confirmed and
// restored"): forcing wireStream's kindID argument to a constant for every node made NO
// test fail, because nothing asserted that a loaded node's streamed KindID column matches
// the kind its spec declared. This drives MoveDispatch.SetNodeStreams (the real,
// exported, production call path — move_streams.go delegates straight into
// streamWiring.setNodeStreams, stream_wiring.go) with a synthetic per-kind kindIDFor
// lookup and asserts each node's OWN nodeGeometry.writeStreamFrame call (one goroutine,
// run synchronously here — no mover goroutine is started) packs back its OWN kindID, not
// some other node's or a constant.
package dispatch

import (
	"context"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

// TestSetNodeStreamsResolvesPerNodeKindID pins that setNodeStreams' kindID column is
// derived, per node, from that node's OWN spec-declared kind (via the injected
// kindIDFor), not a shared/constant value. Two nodes of two DIFFERENT kinds, mapped by
// this test's own kindIDFor to two DIFFERENT, non-zero, non-255 ids (255 is the exact
// wrong constant a prior break forced onto every node — see this file's header), so a
// regression that hardcodes a single value or swaps the two nodes fails this test.
func TestSetNodeStreamsResolvesPerNodeKindID(t *testing.T) {
	const topo = `{
	  "nodes": [
	    {"id":"1","type":"AimedSrc","scenePolarR":0,"scenePolarTheta":0,"scenePolarPhi":0},
	    {"id":"2","type":"AimedSink","scenePolarR":50,"scenePolarTheta":1.5707963267948966,"scenePolarPhi":0}
	  ],
	  "edges": [
	    {"label":"e0","kind":"data","source":"1","sourceHandle":"Out","target":"2","targetHandle":"In"}
	  ]
	}`
	root := writeSpecTree(t, t.TempDir(), topo)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, md, _, err := LoadTopology(ctx, root, T.New(), clock.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology: %v", err)
	}

	// kindIDFor mirrors production's shape (runtopology/node_stream.go passes
	// Buffer.NodeKindID here) but with a synthetic map local to this test, so the
	// expected ids are unambiguous fixture constants rather than depending on
	// Buffer's generated kindIDMap staying stable. Neither value is 0 or 255
	// (KindIDUnknown / the exact constant a prior injected break used), and the two
	// differ from each other, so this test cannot pass by coincidence.
	const (
		aimedSrcKindID  uint8 = 7
		aimedSinkKindID uint8 = 42
	)
	kindIDFor := func(kind string) uint8 {
		switch kind {
		case "AimedSrc":
			return aimedSrcKindID
		case "AimedSink":
			return aimedSinkKindID
		default:
			t.Fatalf("kindIDFor: unexpected kind %q", kind)
			return 0
		}
	}

	// captured records the KindID each node's OWN writeStreamFrame call actually
	// packed into its frame, keyed by NodeRow (row = id-1, persistence-ownership.md).
	captured := map[int32]uint8{}
	buildFrame := func(f nodeactor.NodeFrameInput) []byte {
		captured[f.NodeRow] = f.KindID
		return nil
	}
	buildInteriorFrame := func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte {
		return nil
	}

	// SetNodeStreams is the real exported production entry point (move_streams.go);
	// nodeBase/interiorBase/driveBase are arbitrary — the underlying os.NewFile calls
	// never need a live fd, since fire-and-forget writes on an invalid fd simply fail
	// silently (this bridge's own documented posture), and buildFrame above is called
	// regardless of write success.
	md.SetNodeStreams(1000, 2000, 0, false, md.RT.NodeRowFor, buildFrame, buildInteriorFrame, kindIDFor)

	src, ok := md.mr.nodeGeoms["1"]
	if !ok {
		t.Fatalf("nodeGeoms missing node %q", "1")
	}
	dst, ok := md.mr.nodeGeoms["2"]
	if !ok {
		t.Fatalf("nodeGeoms missing node %q", "2")
	}

	// Trigger each node's OWN emit — the same call its own driving goroutine makes on
	// every tick/change (node_geometry_stream.go's emitGeometry); called here directly,
	// synchronously, on this test's single goroutine (testing-shape.md: "a goroutine's
	// own emitted stream frame" is a legitimate one-goroutine subject).
	src.WriteStreamFrame(nil)
	dst.WriteStreamFrame(nil)

	if got, want := captured[src.NodeRow()], aimedSrcKindID; got != want {
		t.Fatalf("node 1 (AimedSrc) streamed KindID = %d, want %d", got, want)
	}
	if got, want := captured[dst.NodeRow()], aimedSinkKindID; got != want {
		t.Fatalf("node 2 (AimedSink) streamed KindID = %d, want %d", got, want)
	}
}
