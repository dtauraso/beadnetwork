package Overlay

type Edit struct {
	Flag string
}

func DecodeUpdate(payload []byte, attr byte) (Edit, bool) {
	r := NewReader(payload, 0)
	if attr != attrToggle {
		return Edit{}, false
	}
	flagID, err := r.U8()
	if err != nil || int(flagID) >= len(FlagNames) {
		return Edit{}, false
	}
	return Edit{Flag: FlagNames[flagID]}, true
}
