package loadspec

import (
	"strconv"

	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
)

func KindForID(id uint8) (string, bool) {
	for _, k := range B.KnownKinds() {
		if B.NodeKindID(k) == id {
			return k, true
		}
	}
	return "", false
}

func NewNodeID(root string) string {
	return strconv.Itoa(LargestNodeID(root) + 1)
}
