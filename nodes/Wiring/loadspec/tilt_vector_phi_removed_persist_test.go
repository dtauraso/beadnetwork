package loadspec_test

// tilt_vector_phi_removed_persist_test.go — moved bodily from nodes/Wiring/dispatch
// (docs/planning/movedispatch-decomposition.md §34): it exercises only loadspec.LoadTree,
// never MoveDispatch/build.LoadTopology.
//
// TestLoadTreeIgnoresLegacyTopTiltVectorPhiIdx: task/drop-tilt-vector-phi removed φ from
// the tilt-vector model end to end, including position.json's TopTiltVectorPhiIdx field.
// A position.json written by an OLDER build that still carries that key must still LOAD —
// encoding/json silently drops an unrecognized field rather than erroring, and this test
// is the proof that positionfile.JSON has no DisallowUnknownFields call anywhere on this
// path that would turn that into a load failure.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
)

// writeTreeFile writes body to <root>/<rel>, creating any missing parent directories. Local
// to this file — the shared dispatch-package helper of the same name was not moved along
// (no other test in this package needs it).
func writeTreeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", p, err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
}

func TestLoadTreeIgnoresLegacyTopTiltVectorPhiIdx(t *testing.T) {
	root := t.TempDir()
	writeTreeFile(t, root, "nodes/1/meta.json", `{"id":"1","type":"PairNode"}`)
	// A legacy position.json exactly as an older build would have written it: the current
	// theta field PLUS the now-removed phi field.
	writeTreeFile(t, root, "nodes/1/position.json",
		`{"scenePolarR":10,"scenePolarTheta":0,"scenePolarPhi":0,`+
			`"quantITheta":0,"quantIPhi":0,"quantIR":0,`+
			`"stepTheta":0.1,"stepPhi":0.1,"stepR":1,`+
			`"topTiltVectorThetaIdx":5,"topTiltVectorPhiIdx":9}`)

	spec, err := loadspec.LoadTree(root)
	if err != nil {
		t.Fatalf("LoadTree: %v (a legacy topTiltVectorPhiIdx field must not be a load error)", err)
	}
	if len(spec.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(spec.Nodes))
	}
	n := spec.Nodes[0]
	if n.TopTiltVectorThetaIdx == nil || *n.TopTiltVectorThetaIdx != 5 {
		t.Fatalf("want TopTiltVectorThetaIdx=5 read from the legacy file, got %v", n.TopTiltVectorThetaIdx)
	}
}
