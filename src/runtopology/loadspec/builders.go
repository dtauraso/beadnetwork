package loadspec

import (
	NodeBuf "github.com/dtauraso/wirefold/src/Node"
	"strconv"
)

func KindForID(id uint8) (string, bool) {
	for _, k := range NodeBuf.KnownKinds() {
		if NodeBuf.NodeKindID(k) == id {
			return k, true
		}
	}
	return "", false
}

func NewNodeID(root string) string {
	return strconv.Itoa(LargestNodeID(root) + 1)
}
