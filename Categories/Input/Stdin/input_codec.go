package Stdin

func DecodeInputRecord(rec []byte) (StdinMsg, bool) {
	if len(rec) == 0 {
		return StdinMsg{}, false
	}
	switch rec[0] {
	case InKindSave:
		return StdinMsg{Type: "save"}, true
	case InKindRawInput:
		return StdinMsg{Type: "raw-input"}, true
	case InKindEditUpdate:
		return decodeEditUpdate(rec)
	}
	return StdinMsg{}, false
}
