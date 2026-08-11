// fixture_kinds_test.go — SrcNode/SinkNode fixture kinds, duplicated from
// nodes/Wiring/dispatch's own fixture_kinds_test.go (a same-directory-only `init()`
// registration — moving a test file out of dispatch does not carry a sibling file's
// init() with it, docs/planning/movedispatch-decomposition.md §34's own lesson) so
// scene_speed_persist_test.go/scene_lattice_persist_test.go, which need a real
// build.LoadTopology tree, can register these kinds in THIS package's own test binary.

package build_test

import (
	"context"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/build"
	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

// srcNode is a minimal source kind with one paced Out.
type srcNode struct {
	Out *wire.Out
}

func (n *srcNode) Update(ctx context.Context) {
	<-ctx.Done()
}

// sinkNode is a minimal sink kind with one paced In.
type sinkNode struct {
	In *wire.In
}

func (n *sinkNode) Update(ctx context.Context) {
	<-ctx.Done()
}

func init() {
	kindapi.RegisterBuilder("SrcNode",
		[]portwiring.PortSpec{{Name: "Out", Dir: portwiring.PortOut}},
		func(a kindapi.BuildArgs) (wire.Node, error) {
			n := &srcNode{}
			n.Out = a.Out("Out")
			return n, nil
		})
	kindapi.RegisterBuilder("SinkNode",
		[]portwiring.PortSpec{{Name: "In", Dir: portwiring.PortIn}},
		func(a kindapi.BuildArgs) (wire.Node, error) {
			n := &sinkNode{}
			n.In = a.In("In")
			return n, nil
		})
}

// writeTree lays down a minimal directory-tree topology (two nodes + one edge) so
// LoadTopology can build a real MoveDispatch. Duplicated from dispatch's own writeTree
// (nodes/Wiring/dispatch/wire_test_helpers_test.go) for the same cross-directory reason.
func writeTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mk := func(rel, body string) { writeTreeFile(t, root, rel, body) }
	mk("nodes/1/meta.json", `{"id":"1","type":"SrcNode","r":100,"scenePolarR":37.4165738677,"scenePolarTheta":1.00685368543,"scenePolarPhi":1.2490457724}`)
	mk("nodes/2/meta.json", `{"id":"2","type":"SinkNode","r":100,"scenePolarR":87.7496438739,"scenePolarTheta":0.96453035788,"scenePolarPhi":-2.15879893034}`)
	mk("nodes/1/edges/e0.json", `{"label":"e0","kind":"data","sourceHandle":"Out","target":"2","targetHandle":"In"}`)
	return root
}

// loadTreeMD loads root through the real production loader, duplicated from dispatch's own
// loadTreeMD (nodes/Wiring/dispatch/scene_edit_persist_test.go) for the same cross-directory
// reason as writeTree above.
func loadTreeMD(t *testing.T, root string) *dispatch.MoveDispatch {
	t.Helper()
	tr := T.New()
	_, _, md, _, err := build.LoadTopology(context.Background(), root, tr, clock.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology: %v", err)
	}
	return md
}
