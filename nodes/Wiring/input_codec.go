// input_codec.go — BINARY decode of the editor→Go input stream.
//
// The TS→Go bridge is a purely BINARY buffer, symmetric with the Go→TS content
// buffer streamed on fd 3. The webview builds a binary RECORD per message and the
// extension host writes each record FRAMED as [len:u32-LE][record] to Go's stdin.
// This file decodes one record (kind byte + fixed numeric fields + length-prefixed
// UTF-8 sections) back into the SAME stdinMsg the old newline-JSON path produced,
// so applyEdit / HandleRawInput dispatch is UNCHANGED — only the wire
// decode differs.
//
// The record layout is defined ONCE here. The TS side
// (tools/topology-vscode/src/schema/input-layout-gen.ts) is GENERATED from this file's
// InputLayoutFingerprint by tools/gen-node-defs, so it cannot hand-drift from Go.
//
// Numbers are little-endian (matching the content buffer's little-endian encoding). Enum discriminators
// (event kind, hit kind, update entity kind, update attr, overlay flag) are u8 indices
// into the shared orderings. There is NO JSON on the wire: every record is fully numeric.
// The live editor→Go traffic is raw-input, overlays toggle (numeric flag-id), and
// the bare `save` COMMAND (Go persists its OWN authoritative scene state).
// edit-create and edit-delete record kinds were REMOVED end-to-end — no live TS
// sender ever emitted them, and their only trigger (a port-drop gesture calling
// PacedWire.Restore()) unconditionally tore down a live wire's in-flight beads. Their
// kind bytes (20, 21) are left as GAPS below, never renumbered. Only edit-update
// remains.
//
// Kind 3 was inKindResend (removed: the ext host now caches the last stream frame and
// replays it on webview "ready" instead of asking Go to re-emit geometry — see
// runCommand.ts's BuildAndRunRunner.lastSnapshot/getLastSnapshot). Left as an intentional
// GAP rather than renumbered, so no other kind's wire value moves.

package Wiring

import (
	"encoding/binary"
	"errors"
	"math"
	"strings"
)

// InputLayoutFingerprint pins the binary input-record layout. It is the SINGLE source: the
// TS-side INPUT_LAYOUT_FINGERPRINT in input-layout-gen.ts is GENERATED from this string
// (tools/gen-node-defs/input_layout.go), so it always matches by construction — regenerate
// (npm run gen:node-defs) after changing any record kind, field, or enum ordering.
//
// No leading version token: it was a hand-bumped marker nothing parsed (parseFPList reads
// the kinds=/eventKinds=/… tokens, never a version), and record FIELDS aren't listed here
// anyway — so a field-layout change (e.g. removing a hit field) wouldn't move this string.
// The real Go↔TS byte-alignment guard is TestInputFixtureCrossLanguage, which drives a
// record through both encoders and diffs the fields; do not re-add a version pretending to
// be a checked invariant.
//
// INPUT_LAYOUT_FINGERPRINT: kinds=save:4,raw-input:10,edit-update:22 eventKinds=pointerdown,pointermove,pointerup,wheel,home hitKinds=port,handhold,node,edge,torus,empty updateKinds=overlays,clock,distanceGroup,scene,tiltVector updateAttrs=toggle,speed,length,selected,theta,phi,reset,start overlayFlags=tori,scenePoles,nodePoles,selSpherePoles,handholds,labelsGlobal,overlays
const InputLayoutFingerprint = "kinds=save:4,raw-input:10,edit-update:22 eventKinds=pointerdown,pointermove,pointerup,wheel,home hitKinds=port,handhold,node,edge,torus,empty updateKinds=overlays,clock,distanceGroup,scene,tiltVector updateAttrs=toggle,speed,length,selected,theta,phi,reset,start overlayFlags=tori,scenePoles,nodePoles,selSpherePoles,handholds,labelsGlobal,overlays"

// Record kind bytes (first byte of every record).
const (
	// Kinds 1 (resume) and 2 (pause) removed — the play/pause clock gate was deleted
	// end-to-end. Intentional gaps, per house style (never renumber a live wire value).
	// Kind 3 (inKindResend) removed — intentional gap, see comment above.
	inKindSave = 4 // save  — Go persists its OWN scene state (bare command)
	// Kind 5 (inKindFadeToggle) removed — the fade feature was deleted end-to-end.
	inKindRawInput = 10 // raw pointer/wheel/home event
	// Kind 20 (inKindEditCreate) removed — edge creation via edit op was deleted
	// end-to-end. Intentional gap, per house style (never renumber a live wire value).
	// Kind 21 (inKindEditDelete) removed — same removal as above.
	inKindEditUpdate = 22 // edit op=update (entity byte + attr byte + numeric payload)
)

// Update attr indices (positional in IN_UPDATE_ATTRS; the order is pinned across both
// languages by the updateAttrs= token in the fingerprint).
const (
	inOverlayAttrToggle       = 0 // overlays: flip one flag
	inClockAttrSpeed          = 1 // clock: set the playback-speed multiplier
	inDistanceGroupAttrLength = 2 // distanceGroup: set the group's target pair length
	inSceneAttrSelected       = 3 // scene: select a tab from the Go-owned scene tab strip
	inTiltVectorAttrTheta     = 4 // tiltVector: adjust the vector's θ index by one click
	inTiltVectorAttrPhi       = 5 // tiltVector: adjust the vector's φ index by one click
	inTiltVectorAttrReset     = 6 // tiltVector: return both indices to 0 (the RESET button)
	inTiltVectorAttrStart     = 7 // tiltVector: begin the vector exchange from the current angles (the START TILT button)
)

// Enum orderings (u8 index → string), shared with input-layout-gen.ts. All five orderings
// (eventKinds/hitKinds/updateKinds/updateAttrs/overlayFlags) are DERIVED from their token in
// the fingerprint, so a Go-side ordering CANNOT drift from the pinned layout: there is no
// second array to reorder. The chain that keeps both languages in lockstep, per enum:
//
//	Go fingerprint (the source)  →[parseFPList, here]→  Go array
//	   Go fingerprint  →[tools/gen-node-defs/input_layout.go parses the same string]→  TS array
//
// Every link is DERIVED from the one Go fingerprint string — there is no second
// hand-authored copy on either side to fall out of sync. These orderings are WIRE INDICES (a u8
// index is all that crosses the bridge), so an unchecked reorder is a silent
// mis-dispatch — a raycast hit on a node decoding as an edge — with nothing to fail.
var (
	inEventKinds   = parseFPList(InputLayoutFingerprint, "eventKinds=")
	inHitKinds     = parseFPList(InputLayoutFingerprint, "hitKinds=")
	inUpdateKinds  = parseFPList(InputLayoutFingerprint, "updateKinds=")
	inUpdateAttrs  = parseFPList(InputLayoutFingerprint, "updateAttrs=")
	inOverlayFlags = parseFPList(InputLayoutFingerprint, "overlayFlags=")
)

// init fails the process at STARTUP if any enum token is missing from the fingerprint.
// Without this a typo'd/renamed token yields a nil list, enumAt returns "", and every
// record carrying that enum is rejected at runtime — a live input bridge that silently
// does nothing. A malformed layout is a build/boot error, not a quiet degradation
// (mirrors gen-node-defs: a malformed wire prop tag is an error, not a silent skip).
func init() {
	for _, e := range []struct {
		marker string
		list   []string
	}{
		{"eventKinds=", inEventKinds},
		{"hitKinds=", inHitKinds},
		{"updateKinds=", inUpdateKinds},
		{"updateAttrs=", inUpdateAttrs},
		{"overlayFlags=", inOverlayFlags},
	} {
		if len(e.list) == 0 {
			panic("input_codec: INPUT_LAYOUT_FINGERPRINT is missing the " + e.marker + " token — the wire enum orderings derive from it")
		}
	}
}

// parseFPList extracts one space-delimited, comma-separated enum token from the
// fingerprint (e.g. marker "hitKinds=" → ["port","handhold",...]). Returns nil if the
// marker is absent; init() above turns that into a startup panic.
func parseFPList(fp, marker string) []string {
	i := strings.Index(fp, marker)
	if i < 0 {
		return nil
	}
	rest := fp[i+len(marker):]
	if sp := strings.IndexByte(rest, ' '); sp >= 0 {
		rest = rest[:sp]
	}
	return strings.Split(rest, ",")
}

var errShortRecord = errors.New("input record truncated")

// recReader is a little-endian cursor over one deframed record body.
type recReader struct {
	b   []byte
	pos int
}

func (r *recReader) u8() (byte, error) {
	if r.pos+1 > len(r.b) {
		return 0, errShortRecord
	}
	v := r.b[r.pos]
	r.pos++
	return v, nil
}

func (r *recReader) i32() (int32, error) {
	if r.pos+4 > len(r.b) {
		return 0, errShortRecord
	}
	v := int32(binary.LittleEndian.Uint32(r.b[r.pos:]))
	r.pos += 4
	return v, nil
}

func (r *recReader) f64() (float64, error) {
	if r.pos+8 > len(r.b) {
		return 0, errShortRecord
	}
	v := math.Float64frombits(binary.LittleEndian.Uint64(r.b[r.pos:]))
	r.pos += 8
	return v, nil
}

func (r *recReader) boolByte() (bool, error) {
	v, err := r.u8()
	return v != 0, err
}

func enumAt(list []string, i byte) string {
	if int(i) < len(list) {
		return list[i]
	}
	return ""
}

// decodeInputRecord decodes one deframed record body (WITHOUT the [len] frame) into a
// stdinMsg. ok=false means the record was malformed/unknown and must be ignored
// (forward-compatible; mirrors the old json.Unmarshal-error `continue`).
func decodeInputRecord(rec []byte) (stdinMsg, bool) {
	if len(rec) == 0 {
		return stdinMsg{}, false
	}
	r := &recReader{b: rec, pos: 1}
	switch rec[0] {
	case inKindSave:
		return stdinMsg{Type: "save"}, true
	case inKindRawInput:
		ev, ok := decodeRawInput(r)
		if !ok {
			return stdinMsg{}, false
		}
		return stdinMsg{Type: "raw-input", Event: &ev}, true
	case inKindEditUpdate:
		// [entityKind][attr][numeric payload]. entity="overlays" (attr toggle, u8 flag-id).
		kindByte, err1 := r.u8()
		if err1 != nil {
			return stdinMsg{}, false
		}
		entity := enumAt(inUpdateKinds, kindByte)
		attr, err2 := r.u8()
		if err2 != nil {
			return stdinMsg{}, false
		}
		switch entity {
		case "overlays":
			if attr != inOverlayAttrToggle {
				return stdinMsg{}, false
			}
			flagID, err := r.u8()
			if err != nil || int(flagID) >= len(inOverlayFlags) {
				return stdinMsg{}, false
			}
			return stdinMsg{Type: "edit", Op: "update", Kind: "overlays", Attr: "toggle", Flag: inOverlayFlags[flagID]}, true
		case "clock":
			if attr != inClockAttrSpeed {
				return stdinMsg{}, false
			}
			// [u8 speed] — the playback multiplier (0/1/2 from the slider).
			speed, err := r.u8()
			if err != nil {
				return stdinMsg{}, false
			}
			return stdinMsg{Type: "edit", Op: "update", Kind: "clock", Attr: "speed", Num: int(speed)}, true
		case "distanceGroup":
			if attr != inDistanceGroupAttrLength {
				return stdinMsg{}, false
			}
			// [u8 groupIndex][u8 dirUp] — groupIndex indexes distanceGroupOrder (0/1/2);
			// dirUp is 1 for the up arrow (×1.1), 0 for down (÷1.1). Flag carries the
			// direction as a readable string ("up"/"down") rather than adding a second
			// numeric field to stdinMsg — Num already carries the group index.
			groupIdx, errG := r.u8()
			if errG != nil {
				return stdinMsg{}, false
			}
			dirUp, errD := r.u8()
			if errD != nil {
				return stdinMsg{}, false
			}
			dir := "down"
			if dirUp != 0 {
				dir = "up"
			}
			return stdinMsg{Type: "edit", Op: "update", Kind: "distanceGroup", Attr: "length", Num: int(groupIdx), Flag: dir}, true
		case "scene":
			if attr != inSceneAttrSelected {
				return stdinMsg{}, false
			}
			// [u8 tabIndex] — an index into Wiring.SceneTabs, the Go-owned tab strip the
			// VIEW frame carries. Out-of-range indices are rejected by SelectScene, not
			// here: the decoder's job is the byte layout, and the tab list is scene
			// state, not wire state.
			tabIdx, err := r.u8()
			if err != nil {
				return stdinMsg{}, false
			}
			return stdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "selected", Num: int(tabIdx)}, true
		case "tiltVector":
			switch attr {
			case inTiltVectorAttrTheta, inTiltVectorAttrPhi:
				// [u8 nodeRow][u8 dirUp] — nodeRow is the target node's buffer ROW (never
				// its id/name — no sidecar), dirUp is 1 for the up arrow (+1 index), 0 for
				// down (-1). Attr already carries WHICH axis (theta/phi), same shape as
				// distanceGroup's groupIndex+dir payload. Flag carries the axis name so
				// applyUpdateTiltVector (stdin_reader.go) can dispatch on a single string
				// field like every other update handler.
				row, errR := r.u8()
				if errR != nil {
					return stdinMsg{}, false
				}
				dirUp, errD := r.u8()
				if errD != nil {
					return stdinMsg{}, false
				}
				dir := "down"
				if dirUp != 0 {
					dir = "up"
				}
				axis := "theta"
				if attr == inTiltVectorAttrPhi {
					axis = "phi"
				}
				return stdinMsg{Type: "edit", Op: "update", Kind: "tiltVector", Attr: axis, Num: int(row), Flag: dir}, true
			case inTiltVectorAttrReset:
				// [u8 nodeRow] — the RESET button (TiltResetButton.tsx). No direction: a
				// reset always returns both indices to 0, so there is nothing else to
				// carry on the wire.
				row, errR := r.u8()
				if errR != nil {
					return stdinMsg{}, false
				}
				return stdinMsg{Type: "edit", Op: "update", Kind: "tiltVector", Attr: "reset", Num: int(row)}, true
			case inTiltVectorAttrStart:
				// [u8 nodeRow] — the START TILT button (TiltVectorButtons.tsx). No
				// direction: Start never touches an index, it only opens the vector
				// exchange from whatever angles are currently set.
				row, errR := r.u8()
				if errR != nil {
					return stdinMsg{}, false
				}
				return stdinMsg{Type: "edit", Op: "update", Kind: "tiltVector", Attr: "start", Num: int(row)}, true
			}
			return stdinMsg{}, false
		}
		return stdinMsg{}, false
	}
	return stdinMsg{}, false
}

func decodeRawInput(r *recReader) (rawInputMsg, bool) {
	var ev rawInputMsg
	var e error
	f := func() float64 {
		v, err := r.f64()
		if err != nil && e == nil {
			e = err
		}
		return v
	}
	i := func() int {
		v, err := r.i32()
		if err != nil && e == nil {
			e = err
		}
		return int(v)
	}
	b := func() bool {
		v, err := r.boolByte()
		if err != nil && e == nil {
			e = err
		}
		return v
	}
	u := func() byte {
		v, err := r.u8()
		if err != nil && e == nil {
			e = err
		}
		return v
	}

	ev.Kind = enumAt(inEventKinds, u())
	ev.X = f()
	ev.Y = f()
	ev.RectLeft = f()
	ev.RectTop = f()
	ev.RectWidth = f()
	ev.RectHeight = f()
	ev.Button = i()
	ev.Ctrl = b()
	ev.Shift = b()
	ev.Alt = b()
	ev.Meta = b()
	ev.DeltaX = f()
	ev.DeltaY = f()
	ev.Fov = f()
	ev.Hit.Kind = enumAt(inHitKinds, u())
	ev.Hit.IsInput = b()
	ev.Hit.NodeRow = i()
	ev.Hit.PortRow = i()
	ev.Hit.EdgeRow = i()
	if e != nil || ev.Kind == "" || ev.Hit.Kind == "" {
		return ev, false
	}
	return ev, true
}

// --- Encoder (used by Go unit tests; the production encoder is input-layout.ts) ------

type recWriter struct{ b []byte }

func (w *recWriter) u8(v byte)     { w.b = append(w.b, v) }
func (w *recWriter) i32(v int32)   { w.b = binary.LittleEndian.AppendUint32(w.b, uint32(v)) }
func (w *recWriter) f64(v float64) { w.b = binary.LittleEndian.AppendUint64(w.b, math.Float64bits(v)) }
func (w *recWriter) boolByte(v bool) {
	if v {
		w.u8(1)
	} else {
		w.u8(0)
	}
}

func enumIndex(list []string, s string) byte {
	for i, v := range list {
		if v == s {
			return byte(i)
		}
	}
	return 0
}

// encodeControl builds a payload-less control record (save).
func encodeControl(kind byte) []byte { return []byte{kind} }

// encodeOverlaysToggle builds an overlays TOGGLE record (test helper).
func encodeOverlaysToggle(flag string) []byte {
	w := &recWriter{}
	w.u8(inKindEditUpdate)
	w.u8(enumIndex(inUpdateKinds, "overlays"))
	w.u8(inOverlayAttrToggle)
	w.u8(enumIndex(inOverlayFlags, flag))
	return w.b
}

// encodeDistanceGroupAdjust builds a distanceGroup LENGTH record (test helper):
// [22][entityKind=distanceGroup][attr=length][u8 groupIndex][u8 dirUp].
func encodeDistanceGroupAdjust(groupIdx int, dirUp bool) []byte {
	w := &recWriter{}
	w.u8(inKindEditUpdate)
	w.u8(enumIndex(inUpdateKinds, "distanceGroup"))
	w.u8(inDistanceGroupAttrLength)
	w.u8(byte(groupIdx))
	if dirUp {
		w.u8(1)
	} else {
		w.u8(0)
	}
	return w.b
}

// encodeRawInput builds a raw-input record from a rawInputMsg (test helper).
func encodeRawInput(ev rawInputMsg) []byte {
	w := &recWriter{}
	w.u8(inKindRawInput)
	w.u8(enumIndex(inEventKinds, ev.Kind))
	w.f64(ev.X)
	w.f64(ev.Y)
	w.f64(ev.RectLeft)
	w.f64(ev.RectTop)
	w.f64(ev.RectWidth)
	w.f64(ev.RectHeight)
	w.i32(int32(ev.Button))
	w.boolByte(ev.Ctrl)
	w.boolByte(ev.Shift)
	w.boolByte(ev.Alt)
	w.boolByte(ev.Meta)
	w.f64(ev.DeltaX)
	w.f64(ev.DeltaY)
	w.f64(ev.Fov)
	w.u8(enumIndex(inHitKinds, ev.Hit.Kind))
	w.boolByte(ev.Hit.IsInput)
	w.i32(int32(ev.Hit.NodeRow))
	w.i32(int32(ev.Hit.PortRow))
	w.i32(int32(ev.Hit.EdgeRow))
	return w.b
}

// frameRecord wraps a record body with the [len:u32-LE] transport frame.
func frameRecord(rec []byte) []byte {
	return append(binary.LittleEndian.AppendUint32(nil, uint32(len(rec))), rec...)
}
