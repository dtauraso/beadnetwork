// geometry.go — the minimal 3-D vector/segment types used by the port/wire
// primitives (Out.outGeom, PacedWire's beadPlacement, position-stream lerp).
// Moved out of nodes/Wiring/curve_params.go (which keeps the CurveParam*
// constants gen-node-defs reads to emit tools/topology-vscode/src/schema/
// curve-params.ts — those constants are Wiring-owned still; only the plain
// vector math primitive moved here so it can be a leaf, dependency-free type
// shared by nodes/wire and nodes/Wiring).

package wire

// Vec3 is a minimal 3-D vector used by port-geometry math.
type Vec3 struct{ X, Y, Z float64 }

// WireSegment is one edge's straight-line segment from the source OUT-port world
// position to the dest IN-port world position. It is per-edge geometry threaded
// from the loader onto the source Out so each placed bead carries the segment it
// is drawn on. P(t) = Start + t*(End-Start).
type WireSegment struct{ Start, End Vec3 }

func (a Vec3) sub(b Vec3) Vec3 { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func (a Vec3) add(b Vec3) Vec3 { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
func (a Vec3) scale(s float64) Vec3 {
	return Vec3{a.X * s, a.Y * s, a.Z * s}
}

// lerp linearly interpolates between a and b at parameter t.
// P(t) = a + t*(b-a). Used by the position stream to evaluate a bead's position.
func lerp(a, b Vec3, t float64) Vec3 {
	return a.add(b.sub(a).scale(t))
}
