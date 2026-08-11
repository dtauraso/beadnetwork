package Wiring

// refuse_structural_edit_emit_test.go — regression for the write-then-emit split described
// in docs/planning/movedispatch-decomposition.md: refuseStructuralEdit itself no longer
// emits a VIEW frame (it only bumps md.ui.editRefused); each of CreateNode/DeleteNode's 12
// call sites is now responsible for calling md.UI.EmitViewFrame(nil) right after. That is 12
// hand-edits with nothing enforcing them — deleting one produced zero test failures when
// this was checked, which is exactly the hazard this test and its sibling guard close.
//
// A refused structural edit is observable two ways: the ui.editRefused counter, and a VIEW
// frame carrying it (view_stream.go's EditRefused field). This asserts BOTH, on the cheapest
// refusal branch in each method — the !md.ui.sceneEditable check at the top, reachable with
// no real tree on disk.

import (
	"io"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/Wiring/sceneswitch"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

// TestCreateNodeRefusalEmitsViewFrame drives CreateNode down its cheapest refusal branch
// (scene not editable) and asserts both the refusal counter and a VIEW frame emission.
func TestCreateNodeRefusalEmitsViewFrame(t *testing.T) {
	md := &MoveDispatch{
		Scenes: sceneswitch.SceneSwitch{TreeRoot: "does-not-matter", Quit: func() {}},
		UI:     viewstate.UIState{SceneEditable: false},
	}
	frames := 0
	md.UI.SetViewStream(io.Discard, func(tick uint32,
		camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi float32,
		flags viewstate.ViewOverlayFlags,
		dragNodeRow int32,
		_ viewstate.ViewSceneState,
		groupLenTime, groupLenInput, groupLenGate float32,
		speed float32,
		sceneCX, sceneCY, sceneCZ, sceneRadius float32,
		events []wire.RowEvent,
	) []byte {
		frames++
		return nil
	})

	md.CreateNode(0, 0, 0, nil)

	if md.UI.EditRefused != 1 {
		t.Fatalf("CreateNode on a non-editable scene: editRefused = %d, want 1", md.UI.EditRefused)
	}
	if frames != 1 {
		t.Fatalf("CreateNode on a non-editable scene: %d VIEW frames emitted, want 1 — a refusal bumped the counter but the webview was never told", frames)
	}
}

// TestDeleteNodeRefusalEmitsViewFrame is CreateNode's sibling for DeleteNode's own cheapest
// refusal branch.
func TestDeleteNodeRefusalEmitsViewFrame(t *testing.T) {
	md := &MoveDispatch{
		Scenes: sceneswitch.SceneSwitch{TreeRoot: "does-not-matter", Quit: func() {}},
		UI:     viewstate.UIState{SceneEditable: false},
	}
	frames := 0
	md.UI.SetViewStream(io.Discard, func(tick uint32,
		camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi float32,
		flags viewstate.ViewOverlayFlags,
		dragNodeRow int32,
		_ viewstate.ViewSceneState,
		groupLenTime, groupLenInput, groupLenGate float32,
		speed float32,
		sceneCX, sceneCY, sceneCZ, sceneRadius float32,
		events []wire.RowEvent,
	) []byte {
		frames++
		return nil
	})

	md.DeleteNode(0, nil)

	if md.UI.EditRefused != 1 {
		t.Fatalf("DeleteNode on a non-editable scene: editRefused = %d, want 1", md.UI.EditRefused)
	}
	if frames != 1 {
		t.Fatalf("DeleteNode on a non-editable scene: %d VIEW frames emitted, want 1 — a refusal bumped the counter but the webview was never told", frames)
	}
}
