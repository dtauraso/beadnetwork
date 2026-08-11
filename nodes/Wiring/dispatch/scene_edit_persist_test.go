package dispatch_test

// scene_edit_persist_test.go — round-trip test for the FSM-applied node-drag position
// edit persister (meta.json x/y/z): an FSM edit → the debounced writer persists it to
// disk preserving sibling fields → a reload reads it back. The former ring-move anchor
// persister (port json anchorId) is gone — docs/bead-model/channels-not-ports.md, a port has no
// file of its own any more.

import (
	"context"
	"io"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/build"
	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenecamera"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

// TestLoadOverlaysEmitsDefaultsWhenNoPersistedKeys guards the regression where an empty
// scene.json (no overlay keys — everything at its default) made LoadOverlays skip emitting
// entirely, so the buffer streamed all-zero (every overlay OFF) instead of the default-visible
// state. LoadOverlays must ALWAYS stream the resolved state (file data or defaults).
func TestLoadOverlaysEmitsDefaultsWhenNoPersistedKeys(t *testing.T) {
	root := writeTree(t) // no view/scene.json → loadSceneOverlays returns found=false
	md := &dispatch.MoveDispatch{UI: viewstate.UIState{OV: viewstate.DefaultOverlayState()}}
	var kinds []string
	// Decentralized (Step C, memory/feedback_no_single_writer_bridge.md): LoadOverlays writes its own VIEW
	// frame directly via md.emitViewFrame; capture the RowEvent kinds it carries instead of
	// the retired central Trace onEvent hook.
	md.UI.SetViewStream(io.Discard, func(tick uint32,
		camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi float32,
		_ viewstate.ViewOverlayFlags,
		dragNodeRow int32,
		_ viewstate.ViewSceneState,
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
	scenepersist.InstallOverlays(&md.UI, root, tr)
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

// writeTree is dispatch.WriteTree (this file moved to package dispatch_test — see
// wire_test_helpers_test.go's doc comment on writeTree/WriteTree for why the canonical
// copy stays in package dispatch).
func writeTree(t *testing.T) string {
	t.Helper()
	return dispatch.WriteTree(t)
}

func loadTreeMD(t *testing.T, root string) *dispatch.MoveDispatch {
	t.Helper()
	tr := T.New()
	_, _, md, _, err := build.LoadTopology(context.Background(), root, tr, clock.NewRealClock())
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
	viewpersist.EnableEditPersist(&md.Persist, &md.Scenes, &md.MR, root)

	// Flip a visible-sense flag off (tori) and the hidden-sense flag on (labelsGlobal off).
	md.UI.OV.ToggleSceneTori(nil)    // SceneToriVisible: true -> false
	md.UI.OV.ToggleLabelsGlobal(nil) // LabelsGlobalVisible: true -> false
	md.Persist.Overlays().Schedule(md.UI.OV)

	ov, found := scenepersist.LoadSceneOverlays(scenepaths.OverlaysFilePath(root))
	if !found {
		t.Fatalf("loadSceneOverlays found no overlay keys after flush")
	}
	if ov.SceneToriVisible {
		t.Fatalf("sceneToriVisible not persisted as hidden")
	}
	if ov.LabelsGlobalVisible {
		t.Fatalf("labelsGlobalVisible not persisted as hidden")
	}
	// Untouched flag keeps its default (visible).
	if !ov.HandholdsVisible {
		t.Fatalf("handholdsVisible should default visible, got hidden")
	}

	// Seed a fresh dispatch from disk and confirm md.UI.OV is restored.
	fresh := &dispatch.MoveDispatch{UI: viewstate.UIState{OV: viewstate.DefaultOverlayState()}}
	scenepersist.InstallOverlays(&fresh.UI, root, nil)
	if fresh.UI.OV.SceneToriVisible {
		t.Fatalf("LoadOverlays did not restore sceneToriVisible=false")
	}
	if fresh.UI.OV.LabelsGlobalVisible {
		t.Fatalf("LoadOverlays did not restore labelsGlobalVisible=false")
	}
}

// TestOverlaysPersistPreservesCamera: camera.json and overlays.json are separate files (one
// writer each) — writing one must never disturb the other.
func TestOverlaysPersistPreservesCamera(t *testing.T) {
	root := writeTree(t)
	md := loadTreeMD(t, root)
	viewpersist.EnableViewpointPersist(&md.Persist, &md.UI, root)
	md.UI.VP.SetViewpoint(wire.Vec3{X: 1, Y: 2, Z: 3}, 200, geom.Dir{Theta: 0.5, Phi: 1.5}, geom.Dir{Theta: 0.05, Phi: 0.15})
	md.UI.VP.EmitViewpoint(nil)

	viewpersist.EnableEditPersist(&md.Persist, &md.Scenes, &md.MR, root)

	md.UI.OV.ToggleSceneTori(nil)
	md.Persist.Overlays().Schedule(md.UI.OV)

	// Camera survives.
	if _, _, _, _, ok := scenecamera.LoadSceneViewpoint(root); !ok {
		t.Fatalf("cameraPolar clobbered by overlay write")
	}
	// Overlay landed.
	ov, found := scenepersist.LoadSceneOverlays(scenepaths.OverlaysFilePath(root))
	if !found || ov.SceneToriVisible {
		t.Fatalf("overlay not persisted alongside camera (found=%v ov=%+v)", found, ov)
	}
}

// TestEnableEditPersistTrueMonolithicNoTree and TestOverlaysPersistMonolithicForm were
// deleted: they asserted behavior for a monolithic-file topologyPath, a form LoadTopology
// no longer accepts (topo_spec.go) — topologyPath is always the tree root directory. There
// is no longer a second form to test against a first.
