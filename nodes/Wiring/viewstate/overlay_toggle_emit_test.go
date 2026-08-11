package viewstate_test

// overlay_toggle_emit_test.go — regression for the bug where ticking an overlay
// checkbox in the editor flipped the Go-side flag but never told the webview:
// applyUpdate's overlays/toggle handler looked up overlayFlagTraceKind for the
// Trace.Kind* to hand emitViewFrame, and that hand-authored map was missing
// "polarVectors" — the flag flipped in Go and nothing rendered. The flip alone
// (TestOverlayToggleFlips in overlay_gen_test.go) is not sufficient coverage; this
// asserts the frame-builder half of that guarantee (every flag reaches
// ViewOverlayFlags). The applyUpdate/toggle-emits-a-VIEW-frame half moved to
// nodes/Wiring/stdinreader/dispatch_edit_overlay_test.go (§30,
// docs/planning/movedispatch-decomposition.md) alongside applyUpdate itself.
//
// Moved from nodes/Wiring/dispatch (docs/planning/movedispatch-decomposition.md §36):
// exercised only viewstate.UIState.SetViewStream/EmitViewFrame, with MoveDispatch used
// only to wrap a bare UIState field — dropped in favor of a bare *viewstate.UIState.

import (
	"io"
	"reflect"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// TestViewFrameCarriesEveryOverlayFlag closes ViewOverlayFlags' one weakness: it is a
// HAND-WRITTEN struct beside the generated flag vocabulary, so a flag added to
// OVERLAY_FLAG_NAMES without a field here (or with a field nobody assigns in
// emitViewFrame) would stream as zero forever — its toggle would flip Go's state, persist,
// and change nothing on screen, which reads as "the overlay is broken" rather than as
// "a field is missing".
//
// Driven off inputcodec.InOverlayFlags (the fingerprint's own list), it asserts both halves: the
// struct carries exactly as many flags as the vocabulary, and with every flag defaulting
// ON, every field arrives set — a field left unassigned in emitViewFrame is a 0 here.
func TestViewFrameCarriesEveryOverlayFlag(t *testing.T) {
	ui := &viewstate.UIState{OV: viewstate.DefaultOverlayState()}
	var got viewstate.ViewOverlayFlags
	ui.SetViewStream(io.Discard, func(tick uint32,
		camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi float32,
		flags viewstate.ViewOverlayFlags,
		dragNodeRow int32,
		_ viewstate.ViewSceneState,
		groupLenTime, groupLenInput, groupLenGate float32,
		speed float32,
		sceneCX, sceneCY, sceneCZ, sceneRadius float32,
		events []wire.RowEvent,
	) []byte {
		got = flags
		return nil
	})
	ui.EmitViewFrame(nil)

	rv := reflect.ValueOf(got)
	if rv.NumField() != len(inputcodec.InOverlayFlags) {
		t.Fatalf("ViewOverlayFlags has %d fields, but there are %d overlay flags (%v) — a flag was added to OVERLAY_FLAG_NAMES without a field to carry it across the frame builder",
			rv.NumField(), len(inputcodec.InOverlayFlags), inputcodec.InOverlayFlags)
	}
	for i := 0; i < rv.NumField(); i++ {
		if rv.Field(i).Interface().(uint8) != 1 {
			t.Fatalf("ViewOverlayFlags.%s arrived 0 with every overlay defaulting ON — emitViewFrame never assigns it, so this flag streams as off whatever the toggle says",
				rv.Type().Field(i).Name)
		}
	}
}
