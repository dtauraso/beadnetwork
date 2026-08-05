package Wiring

// overlay_toggle_emit_test.go — regression for the bug where ticking an overlay
// checkbox in the editor flipped the Go-side flag but never told the webview:
// applyUpdate's overlays/toggle handler looked up overlayFlagTraceKind for the
// Trace.Kind* to hand emitViewFrame, and that hand-authored map was missing
// "polarVectors" — the flag flipped in Go and nothing rendered. The flip alone
// (TestOverlayToggleFlips in overlay_gen_test.go) is not sufficient coverage; this
// asserts the EMIT that flip is supposed to trigger.

import (
	"io"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

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
	for flag := range overlayToggles {
		flag := flag
		t.Run(flag, func(t *testing.T) {
			wantKind, ok := overlayFlagTraceKind[flag]
			if !ok {
				t.Fatalf("overlay flag %q is in overlayToggles (a live checkbox can send it) but has no entry in overlayFlagTraceKind — its toggle flips the flag and emits nothing to the webview", flag)
			}

			md := &MoveDispatch{ui: uiState{ov: defaultOverlayState()}}
			var kinds []string
			md.SetViewStream(io.Discard, func(tick uint32,
				camPX, camPY, camPZ, camR, camPosTheta, camPosPhi, camUpTheta, camUpPhi float32,
				sceneTori, scenePoles, nodePoles, selSpherePoles, handholds, labelsGlobal, overlaysVis uint8,
				dragNodeRow int32,
				groupLenTime, groupLenInput, groupLenGate float32,
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
			msg := stdinMsg{Type: "edit", Op: "update", Kind: "overlays", Attr: "toggle", Flag: flag}
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
