package Scene

type Edit struct {
	Attr string
	Num  int
	X, Y float64
}

type TiltEdit struct {
	Attr string
	Num  int
	Flag string
}

func DecodeUpdateScene(payload []byte, attr byte) (Edit, bool) {
	r := NewReader(payload, 0)
	switch attr {
	case attrSelected:

		tabIdx, err := r.U8()
		if err != nil {
			return Edit{}, false
		}
		return Edit{Attr: "selected", Num: int(tabIdx)}, true
	case attrLatticePoints:

		points, err := r.U8()
		if err != nil {
			return Edit{}, false
		}
		return Edit{Attr: "latticePoints", Num: int(points)}, true
	case attrCreate:

		kindID, err := r.U8()
		if err != nil {
			return Edit{}, false
		}
		ndcX, err := r.F32()
		if err != nil {
			return Edit{}, false
		}
		ndcY, err := r.F32()
		if err != nil {
			return Edit{}, false
		}
		return Edit{
			Attr: "create",
			Num:  int(kindID), X: float64(ndcX), Y: float64(ndcY),
		}, true
	case attrDelete:

		row, err := r.U8()
		if err != nil {
			return Edit{}, false
		}
		return Edit{Attr: "delete", Num: int(row)}, true
	case attrViewport:

		w, errW := r.F32()
		if errW != nil {
			return Edit{}, false
		}
		h, errH := r.F32()
		if errH != nil {
			return Edit{}, false
		}
		return Edit{Attr: "viewport", X: float64(w), Y: float64(h)}, true
	}
	return Edit{}, false
}

func DecodeUpdateTiltVector(payload []byte, attr byte) (TiltEdit, bool) {
	r := NewReader(payload, 0)
	switch attr {
	case attrTiltVectorPhi:

		row, errR := r.U8()
		if errR != nil {
			return TiltEdit{}, false
		}
		dirUp, errD := r.U8()
		if errD != nil {
			return TiltEdit{}, false
		}
		dir := dirWord(dirUp)
		return TiltEdit{Attr: "phi", Num: int(row), Flag: dir}, true
	case attrTiltVectorRst:

		row, errR := r.U8()
		if errR != nil {
			return TiltEdit{}, false
		}
		return TiltEdit{Attr: "reset", Num: int(row)}, true
	case attrTiltVectorStrt:

		row, errR := r.U8()
		if errR != nil {
			return TiltEdit{}, false
		}
		return TiltEdit{Attr: "start", Num: int(row)}, true
	}
	return TiltEdit{}, false
}

func dirWord(dirUp byte) string {
	if dirUp != 0 {
		return "up"
	}
	return "down"
}
