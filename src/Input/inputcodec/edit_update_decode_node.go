package inputcodec

import "github.com/dtauraso/wirefold/src/Node/Wiring/recread"

func decodeUpdateNode(r *recread.Reader, attr byte) (StdinMsg, bool) {
	switch attr {
	case InNodeAttrDragPhi:
		row, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "dragPhi", Num: int(row)}, true
	case InNodeAttrDragMaxTheta:
		row, errR := r.U8()
		if errR != nil {
			return StdinMsg{}, false
		}
		degrees, errD := r.F32()
		if errD != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "dragMaxTheta", Num: int(row), X: float64(degrees)}, true
	case InNodeAttrDragActive:
		row, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "dragActive", Num: int(row)}, true
	case InNodeAttrKindActive:
		row, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "kindActive", Num: int(row)}, true
	case InNodeAttrSelfDragPhi:
		row, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "selfDragPhi", Num: int(row)}, true
	case InNodeAttrSelfDragMaxTheta:
		row, errR := r.U8()
		if errR != nil {
			return StdinMsg{}, false
		}
		degrees, errD := r.F32()
		if errD != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "selfDragMaxTheta", Num: int(row), X: float64(degrees)}, true
	case InNodeAttrSelfDragActive:
		row, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "selfDragActive", Num: int(row)}, true
	case InNodeAttrDragR:
		row, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "dragR", Num: int(row)}, true
	case InNodeAttrSelfDragR:
		row, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "selfDragR", Num: int(row)}, true
	}
	return StdinMsg{}, false
}
