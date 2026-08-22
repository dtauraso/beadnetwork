package edge

import (
	"github.com/dtauraso/wirefold/src/Input/Codec"
	"github.com/dtauraso/wirefold/src/Input/Stdin"
)

func decodeUpdate(r *Codec.Reader, attr byte) (Stdin.StdinMsg, bool) {
	switch attr {
	case attrDragActive:
		row, err := r.U8()
		if err != nil {
			return Stdin.StdinMsg{}, false
		}
		return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "edge", Attr: "dragActive", Num: int(row)}, true
	}
	return Stdin.StdinMsg{}, false
}

func init() { Stdin.RegisterUpdateDecoder("edge", decodeUpdate) }
