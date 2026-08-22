package Scene

import (
	"github.com/dtauraso/wirefold/Categories/Input/Stdin"
)

func init() {
	Stdin.RegisterUpdateDecoder("scene", decodeUpdateScene)
	Stdin.RegisterUpdateDecoder("tiltVector", decodeUpdateTiltVector)
}

func decodeUpdateScene(payload []byte, attr byte) (Stdin.StdinMsg, bool) {
	r := NewReader(payload, 0)
	switch attr {
	case attrSelected:

		tabIdx, err := r.U8()
		if err != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "selected", Num: int(tabIdx)}, true
	case attrLatticePoints:

		points, err := r.U8()
		if err != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "latticePoints", Num: int(points)}, true
	case attrCreate:

		kindID, err := r.U8()
		if err != nil {
			return Stdin.StdinMsg{}, false
		}
		ndcX, err := r.F32()
		if err != nil {
			return Stdin.StdinMsg{}, false
		}
		ndcY, err := r.F32()
		if err != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{
			Type: "edit", Op: "update", Kind: "scene", Attr: "create",
			Num: int(kindID), X: float64(ndcX), Y: float64(ndcY),
		}, true
	case attrDelete:

		row, err := r.U8()
		if err != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "delete", Num: int(row)}, true
	case attrViewport:

		w, errW := r.F32()
		if errW != nil {
			return Stdin.StdinMsg{}, false
		}
		h, errH := r.F32()
		if errH != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "viewport", X: float64(w), Y: float64(h)}, true
	}
	return Stdin.StdinMsg{}, false
}

func decodeUpdateTiltVector(payload []byte, attr byte) (Stdin.StdinMsg, bool) {
	r := NewReader(payload, 0)
	switch attr {
	case attrTiltVectorPhi:

		row, errR := r.U8()
		if errR != nil {
			return Stdin.StdinMsg{}, false
		}
		dirUp, errD := r.U8()
		if errD != nil {
			return Stdin.StdinMsg{}, false
		}
		dir := Stdin.DirWord(dirUp)
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "tiltVector", Attr: "phi", Num: int(row), Flag: dir}, true
	case attrTiltVectorRst:

		row, errR := r.U8()
		if errR != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "tiltVector", Attr: "reset", Num: int(row)}, true
	case attrTiltVectorStrt:

		row, errR := r.U8()
		if errR != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "tiltVector", Attr: "start", Num: int(row)}, true
	}
	return Stdin.StdinMsg{}, false
}
