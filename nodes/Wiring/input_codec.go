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
// INPUT_LAYOUT_FINGERPRINT: kinds=save:4,raw-input:10,edit-update:22 eventKinds=pointerdown,pointermove,pointerup,wheel,home hitKinds=port,handhold,node,edge,torus,empty updateKinds=overlays,clock,distanceGroup,scene,tiltVector updateAttrs=toggle,speed,length,selected,theta,phi,reset,start,latticePoints,create,delete overlayFlags=tori,scenePoles,nodePoles,selSpherePoles,handholds,labelsGlobal,overlays,nodeBody,nodeRing,ringPick,selectionRing,hoverRing,reachSphere
const InputLayoutFingerprint = "kinds=save:4,raw-input:10,edit-update:22 eventKinds=pointerdown,pointermove,pointerup,wheel,home hitKinds=port,handhold,node,edge,torus,empty updateKinds=overlays,clock,distanceGroup,scene,tiltVector updateAttrs=toggle,speed,length,selected,theta,phi,reset,start,latticePoints,create,delete overlayFlags=tori,scenePoles,nodePoles,selSpherePoles,handholds,labelsGlobal,overlays,nodeBody,nodeRing,ringPick,selectionRing,hoverRing,reachSphere"

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
	// attr 5 (φ) is a GAP — the tilt vector is θ-only end to end now
	// (task/drop-tilt-vector-phi); never renumber the survivors.
	inTiltVectorAttrReset    = 6 // tiltVector: return the index to 0 (the RESET button)
	inTiltVectorAttrStart    = 7 // tiltVector: begin the vector exchange from the current angles (the START TILT button)
	inSceneAttrLatticePoints = 8 // scene: set the pair lattice's point count
	// scene: CREATE a node at a dropped world point, and DELETE the node on a buffer row.
	// ATTRIBUTES, not the create/delete OPS that were removed end-to-end (their kind bytes
	// 20/21 stay gaps and are not reused): the edit surface still has exactly one op, and a
	// new capability is a new entity kind or attribute (CLAUDE.md's bridge-surface rule).
	inSceneAttrCreate = 9  // scene: create a node of a kind, at a world point
	inSceneAttrDelete = 10 // scene: delete the node on a buffer row, and every edge touching it
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

// f32 reads a 4-byte float. The drop point rides as f32, not f64: it is a WORLD POSITION,
// the same precision every position on the content buffer already uses, and it is about to
// be rounded onto the node lattice anyway.
func (r *recReader) f32() (float32, error) {
	if r.pos+4 > len(r.b) {
		return 0, errShortRecord
	}
	v := math.Float32frombits(binary.LittleEndian.Uint32(r.b[r.pos:]))
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
			switch attr {
			case inSceneAttrSelected:
				// [u8 tabIndex] — an index into Wiring.SceneTabs, the Go-owned tab strip the
				// VIEW frame carries. Out-of-range indices are rejected by SelectScene, not
				// here: the decoder's job is the byte layout, and the tab list is scene
				// state, not wire state.
				tabIdx, err := r.u8()
				if err != nil {
					return stdinMsg{}, false
				}
				return stdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "selected", Num: int(tabIdx)}, true
			case inSceneAttrLatticePoints:
				// [u8 points] — the pair lattice's new point count. Out-of-range/non-multiple
				// values are rejected by the handler (applyUpdateScene's "latticePoints"
				// case), not here: the decoder's job is the byte layout only.
				points, err := r.u8()
				if err != nil {
					return stdinMsg{}, false
				}
				return stdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "latticePoints", Num: int(points)}, true
			case inSceneAttrCreate:
				// [u8 kindId][f32 ndcX][f32 ndcY] — the kind to create (its NODE_DEFS id,
				// the same numeric kind identity the Node block's KindId column carries, so
				// no kind NAME crosses this wire) and WHERE ON SCREEN it was dropped, in
				// normalized device coordinates.
				//
				// SCREEN, not world. Turning a drop into a place in the scene needs the
				// camera, and the camera is Go's: the same rayDirThroughNDC every gesture
				// already uses turns this into a world point (scene_structure.go). TS
				// forwards where the pointer was, exactly as raw-input does, and computes
				// no geometry. Which node it connects to is not here either — Go picks the
				// nearest from its own node positions.
				kindID, err := r.u8()
				if err != nil {
					return stdinMsg{}, false
				}
				ndcX, err := r.f32()
				if err != nil {
					return stdinMsg{}, false
				}
				ndcY, err := r.f32()
				if err != nil {
					return stdinMsg{}, false
				}
				return stdinMsg{
					Type: "edit", Op: "update", Kind: "scene", Attr: "create",
					Num: int(kindID), X: float64(ndcX), Y: float64(ndcY),
				}, true
			case inSceneAttrDelete:
				// [u8 nodeRow] — the target's buffer ROW, never its id or name (no sidecar,
				// same as every other addressed edit). Go resolves the row to a node.
				row, err := r.u8()
				if err != nil {
					return stdinMsg{}, false
				}
				return stdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "delete", Num: int(row)}, true
			}
			return stdinMsg{}, false
		case "tiltVector":
			switch attr {
			case inTiltVectorAttrTheta:
				// [u8 nodeRow][u8 dirUp] — nodeRow is the target node's buffer ROW (never
				// its id/name — no sidecar), dirUp is 1 for the up arrow (+1 index), 0 for
				// down (-1). There is only one axis now (theta), so attr alone identifies
				// this as a theta adjust — same shape as distanceGroup's groupIndex+dir
				// payload.
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
				return stdinMsg{Type: "edit", Op: "update", Kind: "tiltVector", Attr: "theta", Num: int(row), Flag: dir}, true
			case inTiltVectorAttrReset:
				// [u8 nodeRow] — the RESET button (TiltResetButton.tsx). No direction: a
				// reset always returns the index to 0, so there is nothing else to
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
