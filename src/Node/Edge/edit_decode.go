package edge

import "github.com/dtauraso/wirefold/src/Input/Codec"

func decodeUpdate(r *Codec.Reader, attr byte) (Codec.StdinMsg, bool) {
	switch attr {
	case Codec.InNodeAttrDragActive:
		row, err := r.U8()
		if err != nil {
			return Codec.StdinMsg{}, false
		}
		return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "edge", Attr: "dragActive", Num: int(row)}, true
	}
	return Codec.StdinMsg{}, false
}

func init() { Codec.RegisterUpdateDecoder("edge", decodeUpdate) }
