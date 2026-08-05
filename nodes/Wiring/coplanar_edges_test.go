package Wiring

import (
	"math"
	"testing"
)

// The whole point: a node's ring PLANE must contain the edge leaving it, so the chain, the
// node's torus and the beads' tori share one plane instead of the chain running through the
// holes. "The plane contains the edge" is exactly "the axis is perpendicular to the edge".
func TestPoleContainingEdgeIsPerpendicularToIt(t *testing.T) {
	self := vec3{X: 100}
	partner := vec3{X: 100, Z: 100} // edge runs along +z
	inT, inP := inwardPole(cart2polar(self))
	theta, phi, ok := poleContainingEdge(inT, inP, self, partner)
	if !ok {
		t.Fatal("a well-separated pair must resolve an axis")
	}
	axis := anglesToWorldOffset(1, theta, phi)
	dir := partner.Sub(self).Normalize()
	if dot := axis.Dot(dir); math.Abs(dot) > 1e-9 {
		t.Fatalf("ring axis must be perpendicular to the edge (the edge lies IN the plane), got dot=%v", dot)
	}
}

// It is the CLOSEST such axis, not an arbitrary perpendicular: the inward pole with its
// along-the-edge component removed. An unrelated perpendicular would tilt the ring further
// from the frame every other scene uses.
func TestPoleContainingEdgeStaysNearestTheInwardPole(t *testing.T) {
	self := vec3{X: 100}
	partner := vec3{X: 100, Z: 100}
	inT, inP := inwardPole(cart2polar(self))
	inward := anglesToWorldOffset(1, inT, inP)
	theta, phi, _ := poleContainingEdge(inT, inP, self, partner)
	axis := anglesToWorldOffset(1, theta, phi)

	dir := partner.Sub(self).Normalize()
	want := inward.Sub(dir.Scale(inward.Dot(dir))).Normalize()
	if axis.Sub(want).Length() > 1e-9 {
		t.Fatalf("axis = %v, want the inward pole projected off the edge = %v", axis, want)
	}
}

// A pole PARALLEL to the edge has no perpendicular component to keep; rather than inventing
// one, report not-resolvable so the caller keeps the axis it had.
func TestPoleParallelToTheEdgeIsNotResolvable(t *testing.T) {
	self := vec3{X: 100}
	partner := vec3{} // edge points straight at the scene centre — along the inward pole
	inT, inP := inwardPole(cart2polar(self))
	if _, _, ok := poleContainingEdge(inT, inP, self, partner); ok {
		t.Fatal("a pole parallel to the edge must report not-resolvable")
	}
}

func TestCoincidentCentresAreNotResolvable(t *testing.T) {
	self := vec3{X: 100}
	if _, _, ok := poleContainingEdge(0, 0, self, self); ok {
		t.Fatal("coincident centres have no edge direction and must report not-resolvable")
	}
}

// Per scene, like the drag: the pair asks for coplanar edges, the ring does not, and an
// unknown tree keeps the plain inward pole.
func TestCoplanarEdgesIsPerScene(t *testing.T) {
	if !SceneWantsCoplanarEdges("topology-pair") {
		t.Fatal("the pair scene must ask for coplanar edges")
	}
	if SceneWantsCoplanarEdges("topology") {
		t.Fatal("the ring scene must keep the plain inward pole")
	}
	if SceneWantsCoplanarEdges("/tmp/some-fixture") {
		t.Fatal("an unknown tree must keep the plain inward pole")
	}
}

// THE RING TAB MUST LOOK EXACTLY AS IT DID. A scene that has not asked for coplanar edges
// streams the torus's OWN normal, which draws as an unrotated ring — the way every ring was
// drawn before ring orientation existed. This is the assertion that keeps a pair-scene
// feature from quietly restyling another scene.
func TestDefaultRingAxisIsTheUnrotatedTorusNormal(t *testing.T) {
	theta, phi := torusDefaultAxisAngles()
	axis := anglesToWorldOffset(1, theta, phi)
	want := vec3{X: 0, Y: 0, Z: 1} // three.js torusGeometry's own normal
	if axis.Sub(want).Length() > 1e-9 {
		t.Fatalf("default ring axis = %v, want the torus's own +Z normal %v — anything else "+
			"silently re-orients every ring in a scene that asked for nothing", axis, want)
	}
}

// UPRIGHT RINGS: the plane must hold BOTH the edge and world +y, so the ring stands up
// along the edge and the node's own up-vector lies IN that plane. An axis of +y itself would
// lie the ring flat and put the vector perpendicular to it — the opposite arrangement, and
// the one this replaced.
func TestUprightRingPlaneHoldsTheEdgeAndUp(t *testing.T) {
	self := vec3{X: 96, Z: -80}
	partner := vec3{X: 96, Z: -40} // the pair's own shape: same height, edge along +z
	theta, phi, ok := uprightRingAxis(self, partner)
	if !ok {
		t.Fatal("a horizontal edge must resolve an upright axis")
	}
	axis := anglesToWorldOffset(1, theta, phi)
	edge := partner.Sub(self).Normalize()
	up := vec3{X: 0, Y: 1, Z: 0}
	if d := axis.Dot(edge); math.Abs(d) > 1e-9 {
		t.Fatalf("the ring plane must contain the EDGE, so the axis is perpendicular to it; got dot=%v", d)
	}
	if d := axis.Dot(up); math.Abs(d) > 1e-9 {
		t.Fatalf("the ring plane must contain UP, so the axis is perpendicular to it; got dot=%v", d)
	}
}

// An edge running straight up has no unique upright plane — every plane through it already
// contains up. Report that rather than returning a degenerate axis.
func TestVerticalEdgeHasNoUniqueUprightPlane(t *testing.T) {
	if _, _, ok := uprightRingAxis(vec3{}, vec3{Y: 100}); ok {
		t.Fatal("an edge parallel to +y must report not-resolvable")
	}
}

// The SECOND vector is a quarter turn from the first, and must stay IN the ring's plane —
// that is the whole point of turning about the ring's own axis rather than any other.
func TestSecondVectorIsAQuarterTurnInsideTheRingPlane(t *testing.T) {
	// The pair's shape: a horizontal edge, so the upright ring plane holds the edge and up.
	self := vec3{X: 96, Z: -80}
	partner := vec3{X: 96, Z: -40}
	axisT, axisP, ok := uprightRingAxis(self, partner)
	if !ok {
		t.Fatal("upright axis must resolve")
	}
	// The first vector points up, and up lies in that plane.
	up := worldDirToAngles(vec3{X: 0, Y: 1, Z: 0})
	upT, upP := up.Theta, up.Phi
	secT, secP := quarterTurnInRingPlane(upT, upP, axisT, axisP)

	axis := anglesToWorldOffset(1, axisT, axisP)
	first := anglesToWorldOffset(1, upT, upP)
	second := anglesToWorldOffset(1, secT, secP)

	if d := second.Dot(axis); math.Abs(d) > 1e-9 {
		t.Fatalf("the second vector must lie IN the ring plane (perpendicular to its axis), got dot=%v", d)
	}
	if d := second.Dot(first); math.Abs(d) > 1e-9 {
		t.Fatalf("a QUARTER turn means perpendicular to the first vector, got dot=%v", d)
	}
	if l := second.Length(); math.Abs(l-1) > 1e-9 {
		t.Fatalf("the turned direction must stay a unit vector, got length %v", l)
	}
}

// The two ends of a pair must point OPPOSITE ways without either being told which side it
// is: each derives its ring axis from its OWN direction to the partner, so the axes are
// already opposites and the same quarter turn lands in opposite world directions.
func TestPairSecondVectorsOppose(t *testing.T) {
	a := vec3{X: 96, Z: -80}
	b := vec3{X: 96, Z: -40}
	up := worldDirToAngles(vec3{X: 0, Y: 1, Z: 0})
	upT, upP := up.Theta, up.Phi

	aAxisT, aAxisP, _ := uprightRingAxis(a, b) // node A measures toward B
	bAxisT, bAxisP, _ := uprightRingAxis(b, a) // node B measures toward A
	aSecT, aSecP := quarterTurnInRingPlane(upT, upP, aAxisT, aAxisP)
	bSecT, bSecP := quarterTurnInRingPlane(upT, upP, bAxisT, bAxisP)

	av := anglesToWorldOffset(1, aSecT, aSecP)
	bv := anglesToWorldOffset(1, bSecT, bSecP)
	if av.Add(bv).Length() > 1e-9 {
		t.Fatalf("the pair's second vectors must be opposites; got %v and %v", av, bv)
	}
}

// The second vector must aim AT the partner, not away from it. This is the assertion the
// earlier pair of tests could not make: perpendicular-to-the-first and opposite-to-each-other
// are both satisfied by the wrong sign too, so only a direction test catches it.
func TestSecondVectorAimsAtThePartner(t *testing.T) {
	a := vec3{X: 96, Z: -80}
	b := vec3{X: 96, Z: -40}
	up := worldDirToAngles(vec3{X: 0, Y: 1, Z: 0})

	aAxisT, aAxisP, _ := uprightRingAxis(a, b)
	aSecT, aSecP := quarterTurnInRingPlane(up.Theta, up.Phi, aAxisT, aAxisP)
	toward := b.Sub(a).Normalize()
	if d := anglesToWorldOffset(1, aSecT, aSecP).Dot(toward); d < 0.999 {
		t.Fatalf("node A's second vector must point AT B: dot with the A→B direction = %v, want ~1", d)
	}

	bAxisT, bAxisP, _ := uprightRingAxis(b, a)
	bSecT, bSecP := quarterTurnInRingPlane(up.Theta, up.Phi, bAxisT, bAxisP)
	back := a.Sub(b).Normalize()
	if d := anglesToWorldOffset(1, bSecT, bSecP).Dot(back); d < 0.999 {
		t.Fatalf("node B's second vector must point back AT A: dot with the B→A direction = %v, want ~1", d)
	}
}
