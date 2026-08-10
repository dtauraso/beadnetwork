// interior_sphere_test.go — the interior beads' TORUS reach must stay inside the node's own
// sphere radius. Split out from the interior package's own node_bead_test.go (god-object
// decomposition): these two assertions need Wiring's own nodeRadius, which package interior
// must not import (Wiring imports interior for InteriorStream/EmitNodeBeads/etc — importing
// back would cycle), so they stay here, calling into interior's exported slot geometry.
package Wiring

import (
	"math"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/interior"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

// TestInteriorBeadsInsideSphere asserts each of the 4 interior bead's TORUS reach
// stays inside the node sphere: |offset| + InteriorTorusOuterR ≤ r. Offsets are
// node-local (centered at origin), so the distance is measured from the origin.
// The torus (outer radius rt), not the sphere, is the bead's true visual extent.
func TestInteriorBeadsInsideSphere(t *testing.T) {
	rt := interior.InteriorTorusOuterR
	r := nodegeom.NodeRadius("Input")
	cases := []struct{ row, col int }{{0, 0}, {0, 1}, {1, 0}, {1, 1}}
	for _, tc := range cases {
		p := interior.InteriorSlotOffset(tc.row, tc.col)
		dist := math.Sqrt(p.X*p.X + p.Y*p.Y + p.Z*p.Z)
		reach := dist + rt
		if reach > r {
			t.Errorf("slot(%d,%d): torus reach %v > r %v — ring pokes outside sphere", tc.row, tc.col, reach, r)
		}
	}
}

// TestInputBeadsInsideSphere asserts the two SelectLeft side beads (at
// ±InteriorSlot on x, vertically centered) keep their torus reach inside the
// node sphere: |offset| + InteriorTorusOuterR ≤ nodeRadius("SelectLeft").
func TestInputBeadsInsideSphere(t *testing.T) {
	rt := interior.InteriorTorusOuterR
	r := nodegeom.NodeRadius("SelectLeft")
	for _, x := range []float64{-interior.InteriorSlot, interior.InteriorSlot} {
		dist := math.Abs(x)
		reach := dist + rt
		if reach > r {
			t.Errorf("side bead x=%v: torus reach %v > r %v — ring pokes outside sphere", x, reach, r)
		}
	}
}
