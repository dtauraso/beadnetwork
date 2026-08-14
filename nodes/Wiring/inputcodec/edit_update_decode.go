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
	case "panels":
		return decodeUpdatePanels(r, attr)
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

func decodeUpdatePanels(r *recread.Reader, attr byte) (StdinMsg, bool) {
	if attr != InPanelAttrToggle {
		return StdinMsg{}, false
	}
	flagID, err := r.U8()
	if err != nil || int(flagID) >= len(InPanelFlags) {
		return StdinMsg{}, false
	}
	return StdinMsg{Type: "edit", Op: "update", Kind: "panels", Attr: "toggle", Flag: InPanelFlags[flagID]}, true
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
