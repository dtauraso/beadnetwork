package clock

import (
	"github.com/dtauraso/wirefold/src/Input/Codec"
	"github.com/dtauraso/wirefold/src/Input/Stdin"
)

func init() { Stdin.RegisterUpdateDecoder("clock", decodeUpdate) }

func decodeUpdate(r *Codec.Reader, attr byte) (Stdin.StdinMsg, bool) {
	if attr != attrSpeed {
		return Stdin.StdinMsg{}, false
	}
	speed, err := r.U8()
	if err != nil {
		return Stdin.StdinMsg{}, false
	}
	return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "clock", Attr: "speed", Num: int(speed)}, true
}
