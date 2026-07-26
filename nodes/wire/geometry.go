// geometry.go — the minimal 3-D vector/segment types used by the port/wire
// primitives (Out.outGeom, PacedWire's beadPlacement, position-stream lerp).
// Moved out of nodes/Wiring/curve_params.go (which keeps the CurveParam*
// constants gen-node-defs reads to emit tools/topology-vscode/src/schema/
// curve-params.ts — those constants are Wiring-owned still; only the plain
// vector math primitive moved here so it can be a leaf, dependency-free type
// shared by nodes/wire and nodes/Wiring).

package wire

import "math"

// Vec3 is a minimal 3-D vector used by port-geometry math. Exported (and its
// methods below) because nodes/Wiring aliases this type as its own `vec3` and
// calls these methods across the package boundary — see
// nodes/Wiring/curve_params.go.
type Vec3 struct{ X, Y, Z float64 }

// WireSegment is one edge's straight-line segment from the source OUT-port world
// position to the dest IN-port world position. It is per-edge geometry threaded
// from the loader onto the source Out so each placed bead carries the segment it
// is drawn on. P(t) = Start + t*(End-Start).
type WireSegment struct{ Start, End Vec3 }

func (a Vec3) Sub(b Vec3) Vec3 { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func (a Vec3) Add(b Vec3) Vec3 { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
func (a Vec3) Scale(s float64) Vec3 {
	return Vec3{a.X * s, a.Y * s, a.Z * s}
}
func (a Vec3) Length() float64 {
	return math.Sqrt(a.X*a.X + a.Y*a.Y + a.Z*a.Z)
}
func (a Vec3) Normalize() Vec3 {
	l := a.Length()
	if l == 0 {
		return Vec3{}
	}
	return Vec3{a.X / l, a.Y / l, a.Z / l}
}
func (a Vec3) Dot(b Vec3) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}
func (a Vec3) Cross(b Vec3) Vec3 {
	return Vec3{
		a.Y*b.Z - a.Z*b.Y,
		a.Z*b.X - a.X*b.Z,
		a.X*b.Y - a.Y*b.X,
	}
}

// lerp linearly interpolates between a and b at parameter t.
// P(t) = a + t*(b-a). Used by the position stream to evaluate a bead's position.
func lerp(a, b Vec3, t float64) Vec3 {
	return a.Add(b.Sub(a).Scale(t))
}
