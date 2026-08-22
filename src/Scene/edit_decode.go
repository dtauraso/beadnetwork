package Scene

import "github.com/dtauraso/wirefold/src/Input/Codec"

func init() {
	Codec.RegisterUpdateDecoder("scene", decodeUpdateScene)
	Codec.RegisterUpdateDecoder("tiltVector", decodeUpdateTiltVector)
}

func decodeUpdateScene(r *Codec.Reader, attr byte) (Codec.StdinMsg, bool) {
	switch attr {
	case Codec.InSceneAttrSelected:

		tabIdx, err := r.U8()
		if err != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "selected", Num: int(tabIdx)}, true
	case Codec.InSceneAttrLatticePoints:

		points, err := r.U8()
		if err != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "latticePoints", Num: int(points)}, true
	case Codec.InSceneAttrCreate:

		kindID, err := r.U8()
		if err != nil {
			return Codec.StdinMsg{}, false
		}
		ndcX, err := r.F32()
		if err != nil {
			return Codec.StdinMsg{}, false
		}
		ndcY, err := r.F32()
		if err != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{
			Type: "edit", Op: "update", Kind: "scene", Attr: "create",
			Num: int(kindID), X: float64(ndcX), Y: float64(ndcY),
		}, true
	case Codec.InSceneAttrDelete:

		row, err := r.U8()
		if err != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "delete", Num: int(row)}, true
	case Codec.InSceneAttrViewport:

		w, errW := r.F32()
		if errW != nil {
			return Codec.StdinMsg{}, false
		}
		h, errH := r.F32()
		if errH != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "scene", Attr: "viewport", X: float64(w), Y: float64(h)}, true
	}
	return Codec.StdinMsg{}, false
}

func decodeUpdateTiltVector(r *Codec.Reader, attr byte) (Codec.StdinMsg, bool) {
	switch attr {
	case Codec.InTiltVectorAttrPhi:

		row, errR := r.U8()
		if errR != nil {
			return Codec.StdinMsg{}, false
		}
		dirUp, errD := r.U8()
		if errD != nil {
			return Codec.StdinMsg{}, false
		}
		dir := Codec.DirWord(dirUp)
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "tiltVector", Attr: "phi", Num: int(row), Flag: dir}, true
	case Codec.InTiltVectorAttrReset:

		row, errR := r.U8()
		if errR != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "tiltVector", Attr: "reset", Num: int(row)}, true
	case Codec.InTiltVectorAttrStart:

		row, errR := r.U8()
		if errR != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "tiltVector", Attr: "start", Num: int(row)}, true
	}
	return Codec.StdinMsg{}, false
}
