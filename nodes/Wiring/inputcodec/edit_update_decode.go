package inputcodec

import "github.com/dtauraso/wirefold/nodes/Wiring/recread"

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

		tabIdx, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "selected", Num: int(tabIdx)}, true
	case InSceneAttrLatticePoints:

		points, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "latticePoints", Num: int(points)}, true
	case InSceneAttrCreate:

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

		row, errR := r.U8()
		if errR != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "tiltVector", Attr: "reset", Num: int(row)}, true
	case InTiltVectorAttrStart:

		row, errR := r.U8()
		if errR != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "tiltVector", Attr: "start", Num: int(row)}, true
	}
	return StdinMsg{}, false
}
