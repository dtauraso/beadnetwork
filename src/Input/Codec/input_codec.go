package Codec

func DecodeInputRecord(rec []byte) (StdinMsg, bool) {
	if len(rec) == 0 {
		return StdinMsg{}, false
	}
	r := NewReader(rec, 1)
	switch rec[0] {
	case InKindSave:
		return StdinMsg{Type: "save"}, true
	case InKindRawInput:
		return StdinMsg{Type: "raw-input"}, true
	case InKindEditUpdate:
		return decodeEditUpdate(r)
	}
	return StdinMsg{}, false
}
