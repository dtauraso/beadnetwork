package loadspec

import (
	"strconv"

	B "github.com/dtauraso/wirefold/Buffer"
)

const (
	VerticalRingNormalX, VerticalRingNormalY, VerticalRingNormalZ = 0.0, 0.0, 1.0
	FlatRingNormalX, FlatRingNormalY, FlatRingNormalZ             = 0.0, 1.0, 0.0
	SideRingNormalX, SideRingNormalY, SideRingNormalZ             = 1.0, 0.0, 0.0
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
