package Wiring

import (
	"testing"
)

// TestNodeTorusOuterR verifies nodeTorusOuterR = nodeRadius(kind) * (1 + ratio),
// the formula chain_beads.go's tangent placement depends on (docs/bead-lattice.md).
func TestNodeTorusOuterR(t *testing.T) {
	for _, kind := range []string{"Input", "Time"} {
		want := nodeRadius(kind) * (1 + ShadingParamNodeRingTubeRatio)
		if got := nodeTorusOuterR(kind); got != want {
			t.Fatalf("nodeTorusOuterR(%q) = %v, want %v", kind, got, want)
		}
	}
}
