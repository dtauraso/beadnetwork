package Node

import "github.com/dtauraso/wirefold/src/Input/Codec"

func decodeUpdate(r *Codec.Reader, attr byte) (Codec.StdinMsg, bool) {
	switch attr {
	case Codec.InNodeAttrDragPhi:
		row, err := r.U8()
		if err != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "dragPhi", Num: int(row)}, true
	case Codec.InNodeAttrDragMaxTheta:
		row, errR := r.U8()
		if errR != nil {
			return Codec.StdinMsg{}, false
		}
		degrees, errD := r.F32()
		if errD != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "dragMaxTheta", Num: int(row), X: float64(degrees)}, true
	case Codec.InNodeAttrDragActive:
		row, err := r.U8()
		if err != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "dragActive", Num: int(row)}, true
	case Codec.InNodeAttrKindActive:
		row, err := r.U8()
		if err != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "kindActive", Num: int(row)}, true
	case Codec.InNodeAttrSelfDragPhi:
		row, err := r.U8()
		if err != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "selfDragPhi", Num: int(row)}, true
	case Codec.InNodeAttrSelfDragMaxTheta:
		row, errR := r.U8()
		if errR != nil {
			return Codec.StdinMsg{}, false
		}
		degrees, errD := r.F32()
		if errD != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "selfDragMaxTheta", Num: int(row), X: float64(degrees)}, true
	case Codec.InNodeAttrSelfDragActive:
		row, err := r.U8()
		if err != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "selfDragActive", Num: int(row)}, true
	case Codec.InNodeAttrDragR:
		row, err := r.U8()
		if err != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "dragR", Num: int(row)}, true
	case Codec.InNodeAttrSelfDragR:
		row, err := r.U8()
		if err != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "selfDragR", Num: int(row)}, true
	}
	return Codec.StdinMsg{}, false
}

func init() { Codec.RegisterUpdateDecoder("node", decodeUpdate) }
