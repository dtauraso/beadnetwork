package inputcodec

import "github.com/dtauraso/wirefold/src/Node/Wiring/recread"

func DecodeInputRecord(rec []byte) (StdinMsg, bool) {
	if len(rec) == 0 {
		return StdinMsg{}, false
	}
	r := recread.NewReader(rec, 1)
	switch rec[0] {
	case InKindSave:
		return StdinMsg{Type: "save"}, true
	case InKindRawInput:
		ev, ok := decodeRawInput(r)
		if !ok {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "raw-input", Event: &ev}, true
	case InKindEditUpdate:
		return decodeEditUpdate(r)
	}
	return StdinMsg{}, false
}
