// dispatch_edit_integration_test.go — tests that drive this package's own stdin dispatch
// (ApplyEdit) using the inputcodec package's real encode/decode. Pure decode/byte-layout
// tests live in inputcodec/input_codec_test.go — this file covers what only THIS package
// decides: applying a decoded message and persisting the result.
//
// Moved here from nodes/Wiring/dispatch (§30, docs/planning/movedispatch-decomposition.md)
// alongside dispatch_edit.go/dispatch_apply.go — this package now imports
// nodes/Wiring/dispatch for *dispatch.MoveDispatch, so its own tests build one through
// dispatch.LoadTopology (exported) rather than the unexported newMoveDispatch constructor
// dispatch's own in-package tests use.

package stdinreader

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/build"
	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"

	T "github.com/dtauraso/wirefold/Trace"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// fixtureSrcNode is a minimal, port-less node kind registered ONLY for this test binary
// (nodes/Wiring/dispatch's own fixture_kinds_test.go registers "SrcNode"/"SinkNode" the
// same way, but that registration lives in a separate test binary — this package's own
// tests need their own since kindapi.BuildRegistry panics on an empty registry rather than
// silently building nothing).
type fixtureSrcNode struct{}

func (n *fixtureSrcNode) Update(ctx context.Context) { <-ctx.Done() }

func init() {
	kindapi.RegisterBuilder("SrcNode", []portwiring.PortSpec{},
		func(a kindapi.BuildArgs) (wire.Node, error) { return &fixtureSrcNode{}, nil })
}

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

// writeMinimalTree lays down a minimal directory-tree topology (one node) so
// dispatch.LoadTopology can build a real *dispatch.MoveDispatch.
func writeMinimalTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "nodes", "1"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	meta := `{"id":"1","type":"SrcNode","r":100,"scenePolarR":37.4165738677,"scenePolarTheta":1.00685368543,"scenePolarPhi":1.2490457724}`
	if err := os.WriteFile(filepath.Join(root, "nodes", "1", "meta.json"), []byte(meta), 0o644); err != nil {
		t.Fatalf("WriteFile meta.json: %v", err)
	}
	return root
}

func loadMinimalMD(t *testing.T, root string) *dispatch.MoveDispatch {
	t.Helper()
	tr := T.New()
	_, _, md, _, err := build.LoadTopology(context.Background(), root, tr, clock.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology: %v", err)
	}
	return md
}

// TestSavePersistsCurrentOverlayState applies an overlays TOGGLE edit (flipping Go's held
// state), then a save, and asserts overlays.json reflects the CURRENT (post-edit) state —
// not a stale/empty snapshot. This is the "Go persists its own current topology" guarantee.
func TestSavePersistsCurrentOverlayState(t *testing.T) {
	root := writeMinimalTree(t)
	md := loadMinimalMD(t, root)
	// overlaysVisible defaults true; toggle flips it to false.
	toggle, ok := inputcodec.DecodeInputRecord(frameOverlaysToggleRecord("overlays"))
	if !ok {
		t.Fatal("decode toggle failed")
	}
	ApplyEdit(context.Background(), toggle, md, nil, nil)
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
