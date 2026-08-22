package clock

import "github.com/dtauraso/wirefold/src/Input/Codec"

func init() { Codec.RegisterUpdateDecoder("clock", decodeUpdate) }

func decodeUpdate(r *Codec.Reader, attr byte) (Codec.StdinMsg, bool) {
	if attr != attrSpeed {
		return Codec.StdinMsg{}, false
	}
	speed, err := r.U8()
	if err != nil {
		return Codec.StdinMsg{}, false
	}
	return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "clock", Attr: "speed", Num: int(speed)}, true
}
