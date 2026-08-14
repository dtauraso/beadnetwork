package inputcodec

import "github.com/dtauraso/wirefold/nodes/Wiring/recread"

func decodeUpdateNode(r *recread.Reader, attr byte) (StdinMsg, bool) {
	switch attr {
	case InNodeAttrOrbitPhi:
		row, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "orbitPhi", Num: int(row)}, true
	case InNodeAttrOrbitMaxTheta:
		row, errR := r.U8()
		if errR != nil {
			return StdinMsg{}, false
		}
		degrees, errD := r.F32()
		if errD != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "orbitMaxTheta", Num: int(row), X: float64(degrees)}, true
	case InNodeAttrOrbitActive:
		row, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "orbitActive", Num: int(row)}, true
	}
	return StdinMsg{}, false
}
