package edge

import (
	"github.com/dtauraso/wirefold/src/Input/Stdin"
)

func decodeUpdate(payload []byte, attr byte) (Stdin.StdinMsg, bool) {
	r := NewReader(payload, 0)
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
