package clock

import (
	"github.com/dtauraso/wirefold/Categories/Input/Stdin"
)

func init() { Stdin.RegisterUpdateDecoder("clock", decodeUpdate) }

func decodeUpdate(payload []byte, attr byte) (Stdin.StdinMsg, bool) {
	r := NewReader(payload, 0)
	if attr != attrSpeed {
		return Stdin.StdinMsg{}, false
	}
	speed, err := r.U8()
	if err != nil {
		return Stdin.StdinMsg{}, false
	}
	return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "clock", Attr: "speed", Num: int(speed)}, true
}
