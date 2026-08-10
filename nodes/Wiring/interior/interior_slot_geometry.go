// interior_slot_geometry.go — the node-local 2x2 interior-bead slot grid.
//
// Split out of port_geometry.go (god-object decomposition, pure move — no logic changes):
// this concern is sizing/placing the interior beads' own grid, independent of a node's
// polar position or a port-to-port segment. Package interior (further god-object
// decomposition, pure move): InteriorSlot/InteriorTorusOuterR/InteriorSlotOffset are
// exported because package Wiring's own sphere-fit tests (interior_sphere_test.go) need
// them alongside nodegeom's own NodeRadius, which this package must not import (NodeRadius
// lives in nodegeom/node_geom.go, part of the node geometry cluster, not this one).
package interior

import wire "github.com/dtauraso/wirefold/nodes/wire"

// vec3 mirrors package Wiring's own local alias (vec_alias.go, mirroring nodegeom's curve_params.go) and package geom's
// (geom/vec3.go) — each package that needs it aliases wire.Vec3 locally rather than
// importing one another for a bare type alias.
type vec3 = wire.Vec3

// Interior bead render dimensions — mirror scene-content.tsx INTERIOR_BEAD_R +
// torus tube fraction; keep in sync. Each interior bead draws a sphere of radius
// interiorBeadR PLUS a torus ring whose OUTER radius is
// interiorBeadR*(1+interiorTorusTubeFrac). The slot pitch must space by the torus
// outer radius (the larger extent), not the sphere, so adjacent rings don't touch.
const (
	interiorBeadR         = 5.0  // sphere radius (INTERIOR_BEAD_R)
	interiorTorusTubeFrac = 0.12 // torus tube fraction; outer radius = r*(1+frac)
	interiorBeadGap       = 0.2  // small gap BETWEEN adjacent toruses
)

// InteriorTorusOuterR is the torus outer radius — the bead's true visual extent.
const InteriorTorusOuterR = interiorBeadR * (1 + interiorTorusTubeFrac) // 5.6

// InteriorSlot is the 2x2 grid half-pitch, computed TORUS-AWARE from the bead's
// torus outer radius plus half the desired inter-torus gap. Adjacent same-row
// beads are 2*InteriorSlot apart, so torus-to-torus gap = 2*InteriorSlot -
// 2*rt = interiorBeadGap. Pitch follows bead size (beads are a fixed visual
// size), NOT the node radius — nodegeom.NodeRadius is used only for the wall-fit guarantee.
const InteriorSlot = InteriorTorusOuterR + interiorBeadGap/2 // 5.9

// InteriorSlotOffset returns the NODE-LOCAL OFFSET of the 2x2 interior grid slot
// at (row, col), relative to the node center (NOT a world position): row 0 =
// top/backup, row 1 = bottom/working; col 0 = left, col 1 = right. The grid is
// sized by the bead's TORUS OUTER RADIUS so adjacent rings keep a small gap and
// never overlap:
//
//	slot   = InteriorTorusOuterR + interiorBeadGap/2
//	dx = (col - 0.5) * 2*slot
//	dy = (0.5 - row) * 2*slot
//	dz = 0
//
// The grid is centered on the node, so offsets are symmetric about (0,0). TS
// renders the bead as a child of the node group, so its world position =
// node center + offset is composed by the scene graph (no node center added on
// the Go side). Discrete — beads snap to these slot centers. The corner bead's
// torus reach (|offset| + rt) must stay inside the node sphere radius r (see
// TestInteriorBeadsInsideSphere, nodes/Wiring's interior_sphere_test.go). The Z
// offset is always 0 (grid is planar).
func InteriorSlotOffset(row, col int) vec3 {
	slot := InteriorSlot
	pitch := 2 * slot
	return vec3{
		X: (float64(col) - 0.5) * pitch,
		Y: (0.5 - float64(row)) * pitch,
		Z: 0,
	}
}
