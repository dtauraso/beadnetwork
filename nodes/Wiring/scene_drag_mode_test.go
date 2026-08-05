// scene_drag_mode_test.go — which drag a scene uses is a property of the scene, and the
// default for anything unrecognised is the quantized one.
//
// This is one goroutine's own decision (the loader's, at build time), not a handoff, so it
// is testable directly — see docs/testing-shape.md.
package Wiring

import (
	"path/filepath"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

func TestSceneDragModeIsPerScene(t *testing.T) {
	for _, tab := range SceneTabs {
		got := SceneUsesQuantizedDrag(filepath.Join("/somewhere", tab.Dir))
		if got != tab.QuantizedDrag {
			t.Fatalf("scene %q: SceneUsesQuantizedDrag = %v, want %v (the tab's own QuantizedDrag)", tab.Name, got, tab.QuantizedDrag)
		}
	}
}

// The pair is the scene that MOTIVATED the split, so assert its answer literally rather
// than only against the table it is read from — a table-only check passes just as happily
// when someone flips the field.
func TestPairSceneDragsContinuously(t *testing.T) {
	if SceneUsesQuantizedDrag("topology-pair") {
		t.Fatalf("the pair scene must drag CONTINUOUSLY: its nodes sit ~40 world units apart while " +
			"one quantized step is wire.BeadStepR, so a step moves the node a large fraction of the " +
			"whole scene and the node cannot be positioned at all")
	}
}

// An unknown tree — every test fixture, every one-off run — keeps the quantized drag, which
// is what every scene did before scenes were selectable. A new scene must OPT IN to the
// continuous drag rather than getting it by being unrecognised.
func TestUnknownSceneKeepsTheQuantizedDrag(t *testing.T) {
	for _, p := range []string{"", "/tmp/some-fixture", "/tmp/topology-pair-copy"} {
		if !SceneUsesQuantizedDrag(p) {
			t.Fatalf("SceneUsesQuantizedDrag(%q) = false; an unrecognised tree must keep the quantized drag", p)
		}
	}
}

// The step size is what makes the two drags different at all: a bead-distance step is
// invisible against a large scene and dominant against a small one. Pin the relationship
// that motivated the split, so a change to BeadStepR that would swallow the pair scene
// shows up here rather than as "I can't drag the nodes".
func TestPairSeparationIsSmallAgainstOneQuantizedStep(t *testing.T) {
	const pairSeparation = 40.3 // node 1 to node 2 in the committed pair scene
	if wire.BeadStepR < pairSeparation/10 {
		return // a step this small would be invisible; the split would no longer be motivated
	}
	// Otherwise the step is a large fraction of the scene — exactly the case the continuous
	// drag exists for, so the pair must not be on the quantized one.
	if SceneUsesQuantizedDrag("topology-pair") {
		t.Fatalf("one quantized step (%.2f) is more than a tenth of the pair's whole extent (%.1f), "+
			"yet the pair is on the quantized drag", wire.BeadStepR, pairSeparation)
	}
}

// The ring is the scene that USED the quantized drag, so assert its answer literally too —
// the table-driven test above passes just as happily when someone flips the field back.
//
// A node is one point; an edge is the distance between two of them; a drag says where the
// point went. The bead count then fills whatever line that leaves, which edgeStepCount
// already derives from the live centre-to-centre distance.
func TestRingSceneDragsContinuously(t *testing.T) {
	if SceneUsesQuantizedDrag("topology") {
		t.Fatalf("the ring scene must drag CONTINUOUSLY: under the quantized drag a node's new " +
			"centre came from the bead operation along the chain axis, so the drag target only " +
			"nominated a direction rather than saying where the node went")
	}
}

// Nothing a user can open is on the quantized drag any more. This is not an assertion that
// it SHOULD stay that way — it is a tripwire, so that whoever adds a scene using it, or
// deletes the path outright, has to come here and say which they meant.
func TestNoCommittedSceneUsesTheQuantizedDrag(t *testing.T) {
	for _, tab := range SceneTabs {
		if tab.QuantizedDrag {
			t.Fatalf("scene %q is on the quantized drag; if that is deliberate, update this test "+
				"and SceneTab.QuantizedDrag's doc comment, which currently records that no "+
				"committed scene uses it", tab.Name)
		}
	}
}
