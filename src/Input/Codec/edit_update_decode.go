package Codec

func decodeEditUpdate(r *Reader) (StdinMsg, bool) {
	kindByte, err1 := r.U8()
	if err1 != nil {
		return StdinMsg{}, false
	}
	entity := EnumAt(InUpdateKinds, kindByte)
	attr, err2 := r.U8()
	if err2 != nil {
		return StdinMsg{}, false
	}
	decode, ok := updateDecoders[entity]
	if !ok {
		return StdinMsg{}, false
	}
	return decode(r, attr)
}

func DirWord(dirUp byte) string {
	if dirUp != 0 {
		return "up"
	}
	return "down"
}
