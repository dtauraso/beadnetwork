package Node

import (
	"github.com/dtauraso/wirefold/Categories/Input/Stdin"
)

func decodeUpdate(payload []byte, attr byte) (Stdin.StdinMsg, bool) {
	r := NewReader(payload, 0)
	switch attr {
	case attrDragPhi:
		row, err := r.U8()
		if err != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "dragPhi", Num: int(row)}, true
	case attrDragMaxTheta:
		row, errR := r.U8()
		if errR != nil {
			return Stdin.StdinMsg{}, false
		}
		degrees, errD := r.F32()
		if errD != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "dragMaxTheta", Num: int(row), X: float64(degrees)}, true
	case attrDragActive:
		row, err := r.U8()
		if err != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "dragActive", Num: int(row)}, true
	case attrKindActive:
		row, err := r.U8()
		if err != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "kindActive", Num: int(row)}, true
	case attrSelfDragPhi:
		row, err := r.U8()
		if err != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "selfDragPhi", Num: int(row)}, true
	case attrSelfDragMaxTheta:
		row, errR := r.U8()
		if errR != nil {
			return Stdin.StdinMsg{}, false
		}
		degrees, errD := r.F32()
		if errD != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "selfDragMaxTheta", Num: int(row), X: float64(degrees)}, true
	case attrSelfDragActive:
		row, err := r.U8()
		if err != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "selfDragActive", Num: int(row)}, true
	case attrDragR:
		row, err := r.U8()
		if err != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "dragR", Num: int(row)}, true
	case attrSelfDragR:
		row, err := r.U8()
		if err != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "node", Attr: "selfDragR", Num: int(row)}, true
	}
	return Stdin.StdinMsg{}, false
}

func init() { Stdin.RegisterUpdateDecoder("node", decodeUpdate) }
