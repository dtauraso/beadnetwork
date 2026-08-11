// builders.go — geometry constants shared by the node/edge movers.
//
// This file used to hold the CENTRAL REFLECTION BUILD PIPELINE: reflectBuild, which took
// an empty struct from wire.Register and filled it in by matching field names and types
// (injectClosures, injectFunc), struct shape (reflectPorts/collectPorts/wirePorts) and
// struct tags (populateData). All of it is gone — every kind now constructs itself through
// Wiring.RegisterBuilder/BuildArgs (build_args.go), so a field the builder forgets is a
// compile error instead of a silent nil at runtime.
//
// Deleting it is what PROVED the migration: if any kind had still depended on reflection,
// removing these functions would not compile.

package loadspec

import (
	"strconv"

	B "github.com/dtauraso/wirefold/Buffer"
)

// VerticalRingNormal and FlatRingNormal are the two great-circle ring normals
// streamed on every node-geometry event so TS never hardcodes ring orientation.
// vertical: ring stands upright (normal points along +Z world axis).
// flat: ring lies flat (normal points along +Y world axis, Three y-up convention).
const (
	VerticalRingNormalX, VerticalRingNormalY, VerticalRingNormalZ = 0.0, 0.0, 1.0
	FlatRingNormalX, FlatRingNormalY, FlatRingNormalZ             = 0.0, 1.0, 0.0
)

// KindForID reverses Buffer's kind-id map: the wire carries the numeric kind identity the
// Node block's KindId column already uses, so no kind NAME crosses the bridge. Moved from
// nodes/Wiring's scene_structure.go (god-object decomposition) — pure over Buffer's own
// KnownKinds/NodeKindID, no Wiring state.
func KindForID(id uint8) (string, bool) {
	for _, k := range B.KnownKinds() {
		if B.NodeKindID(k) == id {
			return k, true
		}
	}
	return "", false
}

// NewNodeID is one past the largest id in root's tree — never a reused hole. Reusing a
// freed id would make a node's identity ambiguous across a session boundary: the same
// directory name would name a different node before and after, which is the whole reason
// ids are not renumbered. Moved from nodes/Wiring's scene_structure.go (god-object
// decomposition) — pure over LargestNodeID, which already lives in this package.
func NewNodeID(root string) string {
	return strconv.Itoa(LargestNodeID(root) + 1)
}
