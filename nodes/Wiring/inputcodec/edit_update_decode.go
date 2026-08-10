// edit_update_decode.go — decode of the ADDRESSED EDIT record (kind byte InKindEditUpdate).
//
// One job: [entityKind][attr][numeric payload] → the StdinMsg that names the entity and the
// attribute. The edit surface has exactly ONE op (update), so the only branching here is
// entity → attribute, one small decoder each; a new capability is a new entity kind or
// attribute (CLAUDE.md's bridge-surface rule), which is a new function in this file.
//
// Every decoder is BYTE LAYOUT ONLY. Range/validity of a decoded value is the handler's job
// (Wiring's stdin_apply.go) — a tab index the scene doesn't have, a lattice point count that
// isn't a multiple — because those are scene state, not wire state. What this file rejects is
// a record that ran out of bytes, or an attr the entity does not carry.

package inputcodec

import "github.com/dtauraso/wirefold/nodes/Wiring/recread"

// decodeEditUpdate decodes an edit-update record's [entityKind][attr][numeric payload] and
// hands the payload to the per-entity decoder. entity="overlays" (attr toggle, u8 flag-id).
func decodeEditUpdate(r *recread.Reader) (StdinMsg, bool) {
	kindByte, err1 := r.U8()
	if err1 != nil {
		return StdinMsg{}, false
	}
	entity := recread.EnumAt(InUpdateKinds, kindByte)
	attr, err2 := r.U8()
	if err2 != nil {
		return StdinMsg{}, false
	}
	switch entity {
	case "overlays":
		return decodeUpdateOverlays(r, attr)
	case "clock":
		return decodeUpdateClock(r, attr)
	case "distanceGroup":
		return decodeUpdateDistanceGroup(r, attr)
	case "scene":
		return decodeUpdateScene(r, attr)
	case "tiltVector":
		return decodeUpdateTiltVector(r, attr)
	}
	return StdinMsg{}, false
}

// dirWord turns an up/down direction byte into the readable word the handlers switch on.
// The direction rides in Flag as a string rather than as a second numeric field on StdinMsg
// — Num already carries the group index / node row.
func dirWord(dirUp byte) string {
	if dirUp != 0 {
		return "up"
	}
	return "down"
}

func decodeUpdateOverlays(r *recread.Reader, attr byte) (StdinMsg, bool) {
	if attr != InOverlayAttrToggle {
		return StdinMsg{}, false
	}
	flagID, err := r.U8()
	if err != nil || int(flagID) >= len(InOverlayFlags) {
		return StdinMsg{}, false
	}
	return StdinMsg{Type: "edit", Op: "update", Kind: "overlays", Attr: "toggle", Flag: InOverlayFlags[flagID]}, true
}

func decodeUpdateClock(r *recread.Reader, attr byte) (StdinMsg, bool) {
	if attr != InClockAttrSpeed {
		return StdinMsg{}, false
	}
	// [u8 speed] — the playback multiplier (0/1/2 from the slider).
	speed, err := r.U8()
	if err != nil {
		return StdinMsg{}, false
	}
	return StdinMsg{Type: "edit", Op: "update", Kind: "clock", Attr: "speed", Num: int(speed)}, true
}

func decodeUpdateDistanceGroup(r *recread.Reader, attr byte) (StdinMsg, bool) {
	if attr != InDistanceGroupAttrLength {
		return StdinMsg{}, false
	}
	// [u8 groupIndex][u8 dirUp] — groupIndex indexes distanceGroupOrder (0/1/2);
	// dirUp is 1 for the up arrow (×1.1), 0 for down (÷1.1). Flag carries the
	// direction as a readable string ("up"/"down") rather than adding a second
	// numeric field to StdinMsg — Num already carries the group index.
	groupIdx, errG := r.U8()
	if errG != nil {
		return StdinMsg{}, false
	}
	dirUp, errD := r.U8()
	if errD != nil {
		return StdinMsg{}, false
	}
	dir := dirWord(dirUp)
	return StdinMsg{Type: "edit", Op: "update", Kind: "distanceGroup", Attr: "length", Num: int(groupIdx), Flag: dir}, true
}

func decodeUpdateScene(r *recread.Reader, attr byte) (StdinMsg, bool) {
	switch attr {
	case InSceneAttrSelected:
		// [u8 tabIndex] — an index into Wiring.SceneTabs, the Go-owned tab strip the
		// VIEW frame carries. Out-of-range indices are rejected by SelectScene, not
		// here: the decoder's job is the byte layout, and the tab list is scene
		// state, not wire state.
		tabIdx, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "selected", Num: int(tabIdx)}, true
	case InSceneAttrLatticePoints:
		// [u8 points] — the pair lattice's new point count. Out-of-range/non-multiple
		// values are rejected by the handler (applyUpdateScene's "latticePoints"
		// case), not here: the decoder's job is the byte layout only.
		points, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "latticePoints", Num: int(points)}, true
	case InSceneAttrCreate:
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
		kindID, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		ndcX, err := r.F32()
		if err != nil {
			return StdinMsg{}, false
		}
		ndcY, err := r.F32()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{
			Type: "edit", Op: "update", Kind: "scene", Attr: "create",
			Num: int(kindID), X: float64(ndcX), Y: float64(ndcY),
		}, true
	case InSceneAttrDelete:
		// [u8 nodeRow] — the target's buffer ROW, never its id or name (no sidecar,
		// same as every other addressed edit). Go resolves the row to a node.
		row, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "delete", Num: int(row)}, true
	}
	return StdinMsg{}, false
}

func decodeUpdateTiltVector(r *recread.Reader, attr byte) (StdinMsg, bool) {
	switch attr {
	case InTiltVectorAttrTheta:
		// [u8 nodeRow][u8 dirUp] — nodeRow is the target node's buffer ROW (never
		// its id/name — no sidecar), dirUp is 1 for the up arrow (+1 index), 0 for
		// down (-1). There is only one axis now (theta), so attr alone identifies
		// this as a theta adjust — same shape as distanceGroup's groupIndex+dir
		// payload.
		row, errR := r.U8()
		if errR != nil {
			return StdinMsg{}, false
		}
		dirUp, errD := r.U8()
		if errD != nil {
			return StdinMsg{}, false
		}
		dir := dirWord(dirUp)
		return StdinMsg{Type: "edit", Op: "update", Kind: "tiltVector", Attr: "theta", Num: int(row), Flag: dir}, true
	case InTiltVectorAttrReset:
		// [u8 nodeRow] — the RESET button (TiltResetButton.tsx). No direction: a
		// reset always returns the index to 0, so there is nothing else to
		// carry on the wire.
		row, errR := r.U8()
		if errR != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "tiltVector", Attr: "reset", Num: int(row)}, true
	case InTiltVectorAttrStart:
		// [u8 nodeRow] — the START TILT button (TiltVectorButtons.tsx). No
		// direction: Start never touches an index, it only opens the vector
		// exchange from whatever angles are currently set.
		row, errR := r.U8()
		if errR != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "tiltVector", Attr: "start", Num: int(row)}, true
	}
	return StdinMsg{}, false
}
