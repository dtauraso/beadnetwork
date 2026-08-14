package inputcodec

import "github.com/dtauraso/wirefold/nodes/Wiring/recread"

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
	case InTiltVectorAttrPhi:

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
