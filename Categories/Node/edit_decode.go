package Node

type Edit struct {
	Attr string
	Num  int
	X    float64
}

func DecodeUpdate(payload []byte, attr byte) (Edit, bool) {
	r := NewReader(payload, 0)
	switch attr {
	case attrDragPhi:
		row, err := r.U8()
		if err != nil {
			return Edit{}, false
		}
		return Edit{Attr: "dragPhi", Num: int(row)}, true
	case attrDragMaxTheta:
		row, errR := r.U8()
		if errR != nil {
			return Edit{}, false
		}
		degrees, errD := r.F32()
		if errD != nil {
			return Edit{}, false
		}
		return Edit{Attr: "dragMaxTheta", Num: int(row), X: float64(degrees)}, true
	case attrDragActive:
		row, err := r.U8()
		if err != nil {
			return Edit{}, false
		}
		return Edit{Attr: "dragActive", Num: int(row)}, true
	case attrKindActive:
		row, err := r.U8()
		if err != nil {
			return Edit{}, false
		}
		return Edit{Attr: "kindActive", Num: int(row)}, true
	case attrSelfDragPhi:
		row, err := r.U8()
		if err != nil {
			return Edit{}, false
		}
		return Edit{Attr: "selfDragPhi", Num: int(row)}, true
	case attrSelfDragMaxTheta:
		row, errR := r.U8()
		if errR != nil {
			return Edit{}, false
		}
		degrees, errD := r.F32()
		if errD != nil {
			return Edit{}, false
		}
		return Edit{Attr: "selfDragMaxTheta", Num: int(row), X: float64(degrees)}, true
	case attrSelfDragActive:
		row, err := r.U8()
		if err != nil {
			return Edit{}, false
		}
		return Edit{Attr: "selfDragActive", Num: int(row)}, true
	case attrDragR:
		row, err := r.U8()
		if err != nil {
			return Edit{}, false
		}
		return Edit{Attr: "dragR", Num: int(row)}, true
	case attrSelfDragR:
		row, err := r.U8()
		if err != nil {
			return Edit{}, false
		}
		return Edit{Attr: "selfDragR", Num: int(row)}, true
	}
	return Edit{}, false
}
