package Stdin

import "github.com/dtauraso/wirefold/src/Input/Codec"

func DecodeInputRecord(rec []byte) (StdinMsg, bool) {
	if len(rec) == 0 {
		return StdinMsg{}, false
	}
	r := Codec.NewReader(rec, 1)
	switch rec[0] {
	case Codec.InKindSave:
		return StdinMsg{Type: "save"}, true
	case Codec.InKindRawInput:
		return StdinMsg{Type: "raw-input"}, true
	case Codec.InKindEditUpdate:
		return decodeEditUpdate(r)
	}
	return StdinMsg{}, false
}
