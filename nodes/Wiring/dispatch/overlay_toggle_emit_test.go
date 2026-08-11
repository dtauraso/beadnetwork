package dispatch

// overlay_toggle_emit_test.go — regression for the bug where ticking an overlay
// checkbox in the editor flipped the Go-side flag but never told the webview:
// applyUpdate's overlays/toggle handler looked up overlayFlagTraceKind for the
// Trace.Kind* to hand emitViewFrame, and that hand-authored map was missing
// "polarVectors" — the flag flipped in Go and nothing rendered. The flip alone
// (TestOverlayToggleFlips in overlay_gen_test.go) is not sufficient coverage; this
// asserts the EMIT that flip is supposed to trigger.

import (
	"io"
	"reflect"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
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
	md := &MoveDispatch{UI: viewstate.UIState{OV: viewstate.DefaultOverlayState()}}
	var got viewstate.ViewOverlayFlags
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
		got = flags
		return nil
	})
	md.UI.EmitViewFrame(nil)

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

// TestApplyUpdateOverlayToggleEmitsViewFrame drives applyUpdate exactly as
// RunStdinReader's dispatch loop does for a top-level edit/update message with
// kind=="overlays" attr=="toggle", and asserts a VIEW frame carrying the flag's
// Trace kind was emitted — not just that the underlying bool flipped.
//
// The flag list iterated here MUST be authoritative and independent of the map
// under test. overlayFlagTraceKind is what this test is checking, so iterating
// its own keys can never notice a MISSING entry — deleting a key just produces
// one fewer subtest, and the loop reports green. overlayToggles is generated
// from the same upstream OVERLAY_FLAG_NAMES source and is the toggle-dispatch
// table itself (a flag reachable here is a flag a live checkbox can send), so
// it is the correct authoritative set to drive from.
func TestApplyUpdateOverlayToggleEmitsViewFrame(t *testing.T) {
	for flag := range viewstate.OverlayToggles {
		flag := flag
		t.Run(flag, func(t *testing.T) {
			wantKind, ok := viewstate.OverlayFlagTraceKind[flag]
			if !ok {
				t.Fatalf("overlay flag %q is in OverlayToggles (a live checkbox can send it) but has no entry in OverlayFlagTraceKind — its toggle flips the flag and emits nothing to the webview", flag)
			}

			md := &MoveDispatch{UI: viewstate.UIState{OV: viewstate.DefaultOverlayState()}}
			var kinds []string
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
			tr.SetSink(io.Discard)
			msg := inputcodec.StdinMsg{Type: "edit", Op: "update", Kind: "overlays", Attr: "toggle", Flag: flag}
			applyUpdate(msg, md, tr, nil)

			seen := false
			for _, k := range kinds {
				if k == wantKind {
					seen = true
					break
				}
			}
			if !seen {
				t.Fatalf("toggling overlay flag %q: no VIEW frame carried Trace kind %q (emitted: %v)", flag, wantKind, kinds)
			}
		})
	}
}
