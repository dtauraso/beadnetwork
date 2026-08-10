package nodegeom

import (
	"testing"
)

// TestNodeTorusOuterR verifies NodeTorusOuterR = NodeRadius(kind) * (1 + ratio),
// the formula chain_beads.go's tangent placement depends on (docs/bead-model/bead-lattice.md).
func TestNodeTorusOuterR(t *testing.T) {
	for _, kind := range []string{"Input", "Time"} {
		want := NodeRadius(kind) * (1 + ShadingParamNodeRingTubeRatio)
		if got := NodeTorusOuterR(kind); got != want {
			t.Fatalf("NodeTorusOuterR(%q) = %v, want %v", kind, got, want)
		}
	}
}
