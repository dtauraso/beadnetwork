package edge

type Edit struct {
	Num int
}

func DecodeUpdate(payload []byte, attr byte) (Edit, bool) {
	r := NewReader(payload, 0)
	switch attr {
	case attrDragActive:
		row, err := r.U8()
		if err != nil {
			return Edit{}, false
		}
		return Edit{Num: int(row)}, true
	}
	return Edit{}, false
}
