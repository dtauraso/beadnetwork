// stdin_input_integration_test.go — tests that drive Wiring's own stdin dispatch
// (ApplyEdit, newMoveDispatch) using the inputcodec package's real encode/decode. Pure
// decode/byte-layout tests live in inputcodec/input_codec_test.go — this file covers what
// only THIS package decides: applying a decoded message and persisting the result.
//
// The RunStdinReader-driving tests (framed partial reads, ctx-cancel shutdown) live in
// nodes/Wiring/stdinreader/stdin_reader_integration_test.go instead: RunStdinReader itself
// moved to its own movemsg-sibling package (stdinreader — framing only), and those two
// tests need no unexported Wiring field, so they run as an external test of that package.
// This file keeps the ONE test that does reach an internal field (md.UI.OV) and so must
// stay an in-package Wiring test — an in-package test of Wiring cannot itself import
// stdinreader, since stdinreader already imports Wiring (a real cycle, not a style choice).

package Wiring

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// enumIndex finds s's position in list — used here to build the exact wire bytes
// inputcodec.DecodeInputRecord expects, without depending on inputcodec's own _test.go
// encode helpers (a _test.go file's symbols are not visible outside its own package).
func enumIndex(list []string, s string) byte {
	for i, v := range list {
		if v == s {
			return byte(i)
		}
	}
	return 0
}

// frameOverlaysToggleRecord builds the exact bytes an overlays-toggle edit record carries
// on the wire: [InKindEditUpdate][entityKind=overlays][InOverlayAttrToggle][flagId].
func frameOverlaysToggleRecord(flag string) []byte {
	return []byte{
		inputcodec.InKindEditUpdate,
		enumIndex(inputcodec.InUpdateKinds, "overlays"),
		inputcodec.InOverlayAttrToggle,
		enumIndex(inputcodec.InOverlayFlags, flag),
	}
}

// TestSavePersistsCurrentOverlayState applies an overlays TOGGLE edit (flipping Go's held
// state), then a save, and asserts overlays.json reflects the CURRENT (post-edit) state —
// not a stale/empty snapshot. This is the "Go persists its own current topology" guarantee.
func TestSavePersistsCurrentOverlayState(t *testing.T) {
	root := t.TempDir()
	md, err := newMoveDispatch(map[string]nodegeom.NodeGeom{}, map[string]inputcodec.EdgeEndpoints{}, nil, nil, nil, clock.NewRealClock(), nil, 0)
	if err != nil {
		t.Fatalf("newMoveDispatch: %v", err)
	}
	// overlaysVisible defaults true; toggle flips it to false.
	toggle, ok := inputcodec.DecodeInputRecord(frameOverlaysToggleRecord("overlays"))
	if !ok {
		t.Fatal("decode toggle failed")
	}
	ApplyEdit(toggle, md, nil, nil)
	if err := scenepersist.WriteSceneOverlays(scenepaths.OverlaysFilePath(root), md.UI.OV); err != nil {
		t.Fatalf("writeSceneOverlays: %v", err)
	}
	raw, _ := os.ReadFile(filepath.Join(root, "view", "overlays.json"))
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("overlays.json invalid: %v", err)
	}
	if string(obj["overlaysActive"]) != "false" {
		t.Fatalf("overlaysActive=%s want false (toggled-off state should persist)", obj["overlaysActive"])
	}
}
