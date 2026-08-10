// input_codec_test.go — decode/round-trip tests for the binary editor→Go input records.
//
// Tests that exercise Wiring-side dispatch (applyEdit, RunStdinReader, newMoveDispatch)
// stay in the Wiring package (stdin_input_integration_test.go) — this file covers only
// what this package itself decides: byte layout and decode.

package inputcodec

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// TestKindsTokenMatchesConstants pins the fingerprint's `kinds=` token to the actual
// InKind* record-byte constants — the Go mirror of input-layout.test.ts's "kinds= token
// matches the actual IN_KIND_* constants" test. On the Go side `kinds=save:4,...` is
// DOCUMENTATION ONLY: parseFPList never reads it (only the eventKinds/hitKinds/updateKinds/
// overlayFlags tokens), so nothing otherwise ties save:4 to InKindSave=4. Without this,
// an InKind* value could drift from the documented byte and still pass the compiler's
// duplicate-case check, yet break the wire (Go encodes the const's value, a peer decodes
// the fingerprint's). The TS side is generated straight from this fingerprint
// (tools/gen-node-defs/input_layout.go), so it cannot itself diverge — but Go's OWN
// kind consts are hand-written literals next to the doc token, and this test is what
// ties those together.
func TestKindsTokenMatchesConstants(t *testing.T) {
	want := map[string]int{
		"save":        InKindSave,
		"raw-input":   InKindRawInput,
		"edit-update": InKindEditUpdate,
	}
	got := map[string]int{}
	for _, entry := range parseFPList(InputLayoutFingerprint, "kinds=") {
		name, valStr, ok := strings.Cut(entry, ":")
		if !ok {
			t.Fatalf("kinds= entry %q is not name:value", entry)
		}
		v, err := strconv.Atoi(valStr)
		if err != nil {
			t.Fatalf("kinds= entry %q has a non-numeric byte: %v", entry, err)
		}
		got[name] = v
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fingerprint kinds= = %v, want (from InKind* consts) %v — a const drifted from the fingerprint (or vice versa)", got, want)
	}
}

func TestDecodeControlRecords(t *testing.T) {
	cases := []struct {
		kind byte
		want string
	}{
		{InKindSave, "save"},
	}
	for _, c := range cases {
		msg, ok := DecodeInputRecord(EncodeControl(c.kind))
		if !ok || msg.Type != c.want {
			t.Fatalf("control kind %d → (%q, ok=%v), want %q", c.kind, msg.Type, ok, c.want)
		}
	}
}

func TestDecodeEditUpdateOverlaysToggle(t *testing.T) {
	// Exact bytes: [22][entityKind=0][attr=toggle=0][flagId(tori)=0].
	rec := EncodeOverlaysToggle("tori")
	if want := []byte{InKindEditUpdate, 0, InOverlayAttrToggle, 0}; !bytes.Equal(rec, want) {
		t.Fatalf("overlays toggle bytes = %v, want %v", rec, want)
	}
	msg, ok := DecodeInputRecord(rec)
	if !ok || msg.Type != "edit" || msg.Op != "update" || msg.Kind != "overlays" || msg.Attr != "toggle" || msg.Flag != "tori" {
		t.Fatalf("overlays toggle decode = %+v ok=%v", msg, ok)
	}
	// A non-tori flag maps by index.
	msg2, _ := DecodeInputRecord(EncodeOverlaysToggle("overlays"))
	if msg2.Flag != "overlays" {
		t.Fatalf("toggle overlays decode flag=%q", msg2.Flag)
	}
}

// TestDecodeEditUpdateDistanceGroupLength exercises the "distance home button" panel's
// wire record: [22][entityKind=distanceGroup][attr=length][u8 groupIndex][u8 dirUp].
func TestDecodeEditUpdateDistanceGroupLength(t *testing.T) {
	rec := EncodeDistanceGroupAdjust(2, true)
	want := []byte{InKindEditUpdate, byte(enumIndex(InUpdateKinds, "distanceGroup")), InDistanceGroupAttrLength, 2, 1}
	if !bytes.Equal(rec, want) {
		t.Fatalf("distanceGroup length bytes = %v, want %v", rec, want)
	}
	msg, ok := DecodeInputRecord(rec)
	if !ok || msg.Type != "edit" || msg.Op != "update" || msg.Kind != "distanceGroup" || msg.Attr != "length" || msg.Num != 2 || msg.Flag != "up" {
		t.Fatalf("distanceGroup length decode = %+v ok=%v", msg, ok)
	}
	msg2, ok2 := DecodeInputRecord(EncodeDistanceGroupAdjust(0, false))
	if !ok2 || msg2.Num != 0 || msg2.Flag != "down" {
		t.Fatalf("distanceGroup length (down) decode = %+v ok=%v", msg2, ok2)
	}
}

// TestDecodeEditUpdateSceneLatticePoints exercises the pair-lattice point-count panel's
// wire record: [22][entityKind=scene][attr=latticePoints][u8 points].
func TestDecodeEditUpdateSceneLatticePoints(t *testing.T) {
	rec := EncodeSceneLatticePoints(12)
	want := []byte{InKindEditUpdate, byte(enumIndex(InUpdateKinds, "scene")), InSceneAttrLatticePoints, 12}
	if !bytes.Equal(rec, want) {
		t.Fatalf("scene latticePoints bytes = %v, want %v", rec, want)
	}
	msg, ok := DecodeInputRecord(rec)
	if !ok || msg.Type != "edit" || msg.Op != "update" || msg.Kind != "scene" || msg.Attr != "latticePoints" || msg.Num != 12 {
		t.Fatalf("scene latticePoints decode = %+v ok=%v", msg, ok)
	}
}

// TestOverlayFlagOrderMatchesFingerprint guards that the derived flag order equals the
// fingerprint's overlayFlags list (self-check on parseOverlayFlags).
func TestOverlayFlagOrderMatchesFingerprint(t *testing.T) {
	want := []string{
		// Scene furniture, drawn around the nodes.
		"tori", "scenePoles", "nodePoles", "selSpherePoles", "handholds", "labelsGlobal", "overlays",
		// The node itself, and what marks the node being touched.
		"nodeBody", "nodeRing", "ringPick", "selectionRing", "hoverRing", "reachSphere",
	}
	if !reflect.DeepEqual(InOverlayFlags, want) {
		t.Fatalf("InOverlayFlags = %v, want %v", InOverlayFlags, want)
	}
}

func TestDecodeRawInputRoundTrip(t *testing.T) {
	in := RawInputMsg{
		Kind: "wheel", X: 12.5, Y: -3.25, RectLeft: 1, RectTop: 2, RectWidth: 800, RectHeight: 600,
		Button: -1, Ctrl: true, Shift: false, Alt: true, Meta: false,
		DeltaX: 4, DeltaY: -8, Fov: 50,
		Hit: RawHit{Kind: "node", IsInput: true, NodeRow: 7, PortRow: -1, EdgeRow: -1},
	}
	msg, ok := DecodeInputRecord(EncodeRawInput(in))
	if !ok || msg.Type != "raw-input" || msg.Event == nil {
		t.Fatalf("raw-input decode failed: ok=%v msg=%+v", ok, msg)
	}
	if !reflect.DeepEqual(*msg.Event, in) {
		t.Fatalf("raw-input round-trip mismatch:\n got  %+v\n want %+v", *msg.Event, in)
	}
}

func TestDecodeTruncatedAndUnknown(t *testing.T) {
	if _, ok := DecodeInputRecord(nil); ok {
		t.Fatal("empty record should not decode")
	}
	if _, ok := DecodeInputRecord([]byte{99}); ok {
		t.Fatal("unknown kind byte should not decode")
	}
	// A truncated overlays-toggle record must be rejected, not panic.
	rec := EncodeOverlaysToggle("tori")
	if _, ok := DecodeInputRecord(rec[:len(rec)-1]); ok {
		t.Fatal("truncated update record should not decode")
	}
}

// TestFrameLenPrefix documents the transport frame is [len:u32-LE][record].
func TestFrameLenPrefix(t *testing.T) {
	rec := EncodeControl(InKindSave)
	frame := FrameRecord(rec)
	if got := binary.LittleEndian.Uint32(frame[:4]); int(got) != len(rec) {
		t.Fatalf("frame length prefix = %d, want %d", got, len(rec))
	}
	if !bytes.Equal(frame[4:], rec) {
		t.Fatal("frame body != record")
	}
}
