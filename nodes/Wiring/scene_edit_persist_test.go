package Wiring

// scene_edit_persist_test.go — round-trip test for the FSM-applied node-drag position
// edit persister (meta.json x/y/z): an FSM edit → the debounced writer persists it to
// disk preserving sibling fields → a reload reads it back. The former ring-move anchor
// persister (port json anchorId) is gone — docs/channels-not-ports.md, a port has no
// file of its own any more.

import (
	"context"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"io"
	"testing"

	T "github.com/dtauraso/wirefold/Trace"
)

// TestLoadOverlaysEmitsDefaultsWhenNoPersistedKeys guards the regression where an empty
// scene.json (no overlay keys — everything at its default) made LoadOverlays skip emitting
// entirely, so the buffer streamed all-zero (every overlay OFF) instead of the default-visible
// state. LoadOverlays must ALWAYS stream the resolved state (file data or defaults).
func TestLoadOverlaysEmitsDefaultsWhenNoPersistedKeys(t *testing.T) {
	root := writeTree(t) // no view/scene.json → loadSceneOverlays returns found=false
	md := &MoveDispatch{ui: uiState{ov: defaultOverlayState()}}
	var kinds []string
	// Decentralized (Step C, memory/feedback_no_single_writer_bridge.md): LoadOverlays writes its own VIEW
	// frame directly via md.emitViewFrame; capture the RowEvent kinds it carries instead of
	// the retired central Trace onEvent hook.
	md.SetViewStream(io.Discard, func(tick uint32,
		camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi float32,
		sceneTori, scenePoles, nodePoles, selSpherePoles, handholds, labelsGlobal, overlaysVis uint8,
		dragNodeRow int32,
		groupLenTime, groupLenInput, groupLenGate float32,
		speed float32,
		sceneCX, sceneCY, sceneCZ, sceneRadius float32,
		events []wire.RowEvent,
	) []byte {
		for _, e := range events {
			kinds = append(kinds, e.Kind)
		}
		return nil
	})
	tr := T.New()
	md.LoadOverlays(root, tr)
	// The default-visible overlay flags must have been emitted, not skipped.
	for _, want := range []string{"scene-tori", "overlays-vis"} {
		seen := false
		for _, k := range kinds {
			if k == want {
				seen = true
				break
			}
		}
		if !seen {
			t.Fatalf("LoadOverlays emitted no %q event (emitted: %v) — an empty scene.json must still stream the default overlay state", want, kinds)
		}
	}
}

// writeTree lays down a minimal directory-tree topology (two nodes + one edge) so
// LoadTopology can build a real MoveDispatch. Positions come from meta.json scenePolar.
func writeTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mk := func(rel, body string) { writeTreeFile(t, root, rel, body) }
	mk("nodes/1/meta.json", `{"id":"1","type":"SrcNode","r":100,"scenePolarR":37.4165738677,"scenePolarTheta":1.00685368543,"scenePolarPhi":1.2490457724}`)
	mk("nodes/2/meta.json", `{"id":"2","type":"SinkNode","r":100,"scenePolarR":87.7496438739,"scenePolarTheta":0.96453035788,"scenePolarPhi":-2.15879893034}`)
	mk("nodes/1/edges/e0.json", `{"label":"e0","kind":"data","sourceHandle":"Out","target":"2","targetHandle":"In"}`)
	return root
}

func loadTreeMD(t *testing.T, root string) *MoveDispatch {
	t.Helper()
	tr := T.New()
	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology: %v", err)
	}
	return md
}

// TestPersistOverlaysRoundTrips: toggle an overlay flag → debounced flush → scene.json carries
// the (inverted) key; a fresh MoveDispatch.LoadOverlays reads it back into md.ui.ov.
func TestPersistOverlaysRoundTrips(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	md.EnableEditPersist(root)

	// Flip a visible-sense flag off (tori) and the hidden-sense flag on (labelsGlobal off).
	md.ToggleSceneTori(nil)    // sceneToriVisible: true -> false
	md.ToggleLabelsGlobal(nil) // labelsGlobalVisible: true -> false
	md.persist.overlays.schedule(md.ui.ov)

	ov, found := loadSceneOverlays(overlaysFilePath(root))
	if !found {
		t.Fatalf("loadSceneOverlays found no overlay keys after flush")
	}
	if ov.sceneToriVisible {
		t.Fatalf("sceneToriVisible not persisted as hidden")
	}
	if ov.labelsGlobalVisible {
		t.Fatalf("labelsGlobalVisible not persisted as hidden")
	}
	// Untouched flag keeps its default (visible).
	if !ov.handholdsVisible {
		t.Fatalf("handholdsVisible should default visible, got hidden")
	}

	// Seed a fresh dispatch from disk and confirm md.ui.ov is restored.
	fresh := &MoveDispatch{ui: uiState{ov: defaultOverlayState()}}
	fresh.LoadOverlays(root, nil)
	if fresh.ui.ov.sceneToriVisible {
		t.Fatalf("LoadOverlays did not restore sceneToriVisible=false")
	}
	if fresh.ui.ov.labelsGlobalVisible {
		t.Fatalf("LoadOverlays did not restore labelsGlobalVisible=false")
	}
}

// TestOverlaysPersistPreservesCamera: camera.json and overlays.json are separate files (one
// writer each) — writing one must never disturb the other.
func TestOverlaysPersistPreservesCamera(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	md.EnableViewpointPersist(root)
	md.SetViewpoint(vec3{X: 1, Y: 2, Z: 3}, 200, dir{Theta: 0.5, Phi: 1.5}, dir{Theta: 0.05, Phi: 0.15})
	md.EmitViewpoint(nil)

	md.EnableEditPersist(root)

	md.ToggleSceneTori(nil)
	md.persist.overlays.schedule(md.ui.ov)

	// Camera survives.
	if _, _, _, _, ok := loadSceneViewpoint(root); !ok {
		t.Fatalf("cameraPolar clobbered by overlay write")
	}
	// Overlay landed.
	ov, found := loadSceneOverlays(overlaysFilePath(root))
	if !found || ov.sceneToriVisible {
		t.Fatalf("overlay not persisted alongside camera (found=%v ov=%+v)", found, ov)
	}
}

// TestEnableEditPersistTrueMonolithicNoTree and TestOverlaysPersistMonolithicForm were
// deleted: they asserted behavior for a monolithic-file topologyPath, a form LoadTopology
// no longer accepts (topo_spec.go) — topologyPath is always the tree root directory. There
// is no longer a second form to test against a first.
