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
