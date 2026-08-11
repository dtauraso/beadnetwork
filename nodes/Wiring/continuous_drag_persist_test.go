// continuous_drag_persist_test.go — a drag is written down in BOTH drag modes.
//
// This is the persistence exception in docs/process/testing-shape.md: bytes on disk, through the
// real writer, rather than an assertion about a field.
//
// The regression it pins is a quiet one. Persisting used to sit inside the quantized
// branch, so a scene on the continuous drag moved, drew, and fanned to its neighbours
// entirely correctly — and lost the position on the next load. Nothing failed at the time;
// the only symptom arrived a reload later.
package Wiring

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/positionfile"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// dragAndReadBack drags a node under the given layout mode and returns what landed in that
// node's own position.json, or "" if nothing was written at all. The mover is built bare
// with a persistRoot, the same shape flush_pending_persists_test.go uses to drive the real
// writer without a loader.
func dragAndReadBack(t *testing.T, quantized bool, target vec3) string {
	t.Helper()
	root := t.TempDir()

	md := &MoveDispatch{}
	md.lq.quantizedLayout = quantized
	md.UI.SceneSphere = geom.SceneSphere{Center: vec3{}, Radius: 100}
	md.mr.nodeGeoms = map[string]*nodeactor.NodeGeometry{}
	md.mr.edgeMovers = map[string]*edgemover.EdgeMover{}
	md.mr.centerMirror = map[string]vec3{}

	nm := nodeactor.NewNodeGeometry("1",
		nodegeom.NodeGeom{NodeIdentity: nodegeom.NodeIdentity{Kind: "PairNode"}, ScenePolar: geom.Cart2polar(vec3{X: 100}), HasPos: true},
		nil, clock.NewRealClock())
	nm.SetPersistRoot(root)
	md.mr.nodeGeoms["1"] = nm

	md.lq.commitNodeMoveLocal(&md.mr, &md.UI, nm, target)

	b, err := os.ReadFile(positionfile.FilePath(root, "1"))
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("read position.json: %v", err)
	}
	return string(b)
}

func TestContinuousDragIsPersisted(t *testing.T) {
	got := dragAndReadBack(t, false, vec3{X: 120, Y: 0, Z: 30})
	if got == "" {
		t.Fatal("a continuous drag wrote NO position.json: the node moves and draws correctly, " +
			"then reverts on the next load — the failure that has no symptom until a reload")
	}
	var pos map[string]any
	if err := json.Unmarshal([]byte(got), &pos); err != nil {
		t.Fatalf("position.json is not valid JSON: %v (%s)", err, got)
	}
	if _, ok := pos["scenePolarR"]; !ok {
		t.Fatalf("position.json carries no scenePolarR — the position SOURCE is missing: %s", got)
	}
}

func TestQuantizedDragIsStillPersisted(t *testing.T) {
	if got := dragAndReadBack(t, true, vec3{X: 120, Y: 0, Z: 30}); got == "" {
		t.Fatal("the quantized drag stopped persisting")
	}
}
