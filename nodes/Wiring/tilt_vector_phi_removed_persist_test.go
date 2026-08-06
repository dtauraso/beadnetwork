package Wiring

import "testing"

// TestLoadTreeIgnoresLegacyTopTiltVectorPhiIdx: task/drop-tilt-vector-phi removed φ from
// the tilt-vector model end to end, including position.json's TopTiltVectorPhiIdx field.
// A position.json written by an OLDER build that still carries that key must still LOAD —
// encoding/json silently drops an unrecognized field rather than erroring, and this test
// is the proof that positionFileJSON has no DisallowUnknownFields call anywhere on this
// path that would turn that into a load failure.
func TestLoadTreeIgnoresLegacyTopTiltVectorPhiIdx(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "nodes/1/meta.json", `{"id":"1","type":"Node1"}`)
	// A legacy position.json exactly as an older build would have written it: the current
	// theta field PLUS the now-removed phi field.
	writeTreeFile(t, root, "nodes/1/position.json",
		`{"scenePolarR":10,"scenePolarTheta":0,"scenePolarPhi":0,`+
			`"quantITheta":0,"quantIPhi":0,"quantIR":0,`+
			`"stepTheta":0.1,"stepPhi":0.1,"stepR":1,`+
			`"topTiltVectorThetaIdx":5,"topTiltVectorPhiIdx":9}`)

	spec, err := loadTree(root)
	if err != nil {
		t.Fatalf("loadTree: %v (a legacy topTiltVectorPhiIdx field must not be a load error)", err)
	}
	if len(spec.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(spec.Nodes))
	}
	n := spec.Nodes[0]
	if n.TopTiltVectorThetaIdx == nil || *n.TopTiltVectorThetaIdx != 5 {
		t.Fatalf("want TopTiltVectorThetaIdx=5 read from the legacy file, got %v", n.TopTiltVectorThetaIdx)
	}
}
