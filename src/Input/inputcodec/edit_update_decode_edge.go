package inputcodec

import "github.com/dtauraso/wirefold/src/Input/recread"

func decodeUpdateEdge(r *recread.Reader, attr byte) (StdinMsg, bool) {
	switch attr {
	case InNodeAttrDragActive:
		row, err := r.U8()
		if err != nil {
			return StdinMsg{}, false
		}
		return StdinMsg{Type: "edit", Op: "update", Kind: "edge", Attr: "dragActive", Num: int(row)}, true
	}
	return StdinMsg{}, false
}
