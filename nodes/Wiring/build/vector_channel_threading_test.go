// vector_channel_threading_test.go — external test package (so it can import PairNode,
// the only kind that asks for a tilt-vector channel, without an import cycle back into
// Wiring) closing the third leg of the load-path coverage gap recorded in
// docs/planning/movedispatch-decomposition.md: buildFromSpec's
// `b.vectorOutByNode, b.vectorInByNode = topoderive.AllocateVectorChannels(b.spec)`
// assignment threads through buildNodes -> BuildArgs.VectorOut/VectorIn -> each PairNode's
// own n.vec.VectorOut/VectorIn, and nothing asserted that threading actually lands each end
// on the right node. Reads those two unexported-but-otherwise-untouchable fields via
// reflection (read-only, before any goroutine is started — LoadTopology never launches one)
// rather than adding a production accessor. Moved bodily from nodes/Wiring/dispatch
// (docs/planning/movedispatch-decomposition.md §34): it exercises build.LoadTopology, never
// MoveDispatch's own methods.
package build_test

import (
	"context"
	"reflect"
	"testing"
	"unsafe"

	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
	_ "github.com/dtauraso/wirefold/nodes/PairNode"
	Bld "github.com/dtauraso/wirefold/nodes/Wiring/build"
)

// unexportedField reads any unexported field by NAME off v (a struct or pointer-to-struct)
// via reflection. Read-only, and only ever called before any goroutine is started
// (LoadTopology never launches one), so there is no concurrent writer to race.
func unexportedField(v reflect.Value, name string) reflect.Value {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	f := v.FieldByName(name)
	return reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
}

// pairID reads node's own n.plumb.PairID (exported field on an unexported plumb struct) —
// this node's spec id, the only way to identify WHICH node a []wire.Node entry is from
// outside the PairNode package (the interface itself carries no id).
func pairID(t *testing.T, node any) int32 {
	t.Helper()
	v := reflect.ValueOf(node)
	plumb := unexportedField(v, "plumb")
	id := plumb.FieldByName("PairID")
	if !id.IsValid() {
		t.Fatalf("node type %T has no plumb.PairID", node)
	}
	return int32(id.Int())
}

// vectorChanOf reads node's own unexported vec.VectorOut/vec.VectorIn field (a
// tiltvector.TiltVectorMsg channel, direction selected by field) via reflection.
func vectorChanOf(t *testing.T, node any, field string) any {
	t.Helper()
	v := reflect.ValueOf(node)
	vec := unexportedField(v, "vec")
	chField := vec.FieldByName(field)
	if !chField.IsValid() {
		t.Fatalf("vec has no field %q", field)
	}
	chField = reflect.NewAt(chField.Type(), unsafe.Pointer(chField.UnsafeAddr())).Elem()
	return chField.Interface()
}

// TestPairNodeVectorChannelsThreadSourceOutTargetIn pins the FULL load-time threading of
// allocateVectorChannels' two returned maps through buildFromSpec, BuildArgs, and each
// PairNode's own build func: the edge's SOURCE node's own n.vec.VectorOut and the TARGET
// node's own n.vec.VectorIn must be the SAME underlying channel. Fails under a swap of
// allocateVectorChannels' two return values at the buildFromSpec call site: the source
// node's VectorOut comes back nil (it was wired from vectorInByNode, which has no entry
// keyed by the source id) instead of matching the target's VectorIn.
func TestPairNodeVectorChannelsThreadSourceOutTargetIn(t *testing.T) {
	const topo = `{
	  "nodes": [
	    {"id":"1","type":"PairNode"},
	    {"id":"2","type":"PairNode"}
	  ],
	  "edges": [
	    {"label":"e0","kind":"data","source":"1","sourceHandle":"Out","target":"2","targetHandle":"In"}
	  ]
	}`
	root := writeSpecTree(t, t.TempDir(), topo)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nodes, _, _, _, err := Bld.LoadTopology(ctx, root, T.New(), clock.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology: %v", err)
	}

	var srcOut chan<- tiltvector.TiltVectorMsg
	var dstIn <-chan tiltvector.TiltVectorMsg
	for _, n := range nodes {
		switch pairID(t, n) {
		case 1: // the edge's SOURCE id ("1" in the fixture above)
			if v, ok := vectorChanOf(t, n, "VectorOut").(chan<- tiltvector.TiltVectorMsg); ok {
				srcOut = v
			}
		case 2: // the edge's TARGET id ("2" in the fixture above)
			if v, ok := vectorChanOf(t, n, "VectorIn").(<-chan tiltvector.TiltVectorMsg); ok {
				dstIn = v
			}
		}
	}

	if srcOut == nil {
		t.Fatalf("source node's (id 1) own VectorOut is nil — source end never wired")
	}
	if dstIn == nil {
		t.Fatalf("target node's (id 2) own VectorIn is nil — target end never wired")
	}

	// Compare identity through the channel's element type via reflect (VectorOut is
	// chan<-, VectorIn is <-chan — different static types over the same underlying
	// channel, so reflect.Value.Pointer() is what actually compares them).
	if reflect.ValueOf(srcOut).Pointer() != reflect.ValueOf(dstIn).Pointer() {
		t.Fatalf("source node's VectorOut and target node's VectorIn are different channels; want the same directed edge channel")
	}
}
