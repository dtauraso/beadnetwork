package Wiring

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// writeTreeFile writes body to <root>/<rel>, creating any missing parent directories.
// It is the package's single fixture-file writer for directory-tree topology test
// fixtures (nodes/<id>/meta.json, nodes/<id>/inputs|outputs/<name>.json,
// nodes/<id>/edges/<label>.json — adjacency layout, edges live under their SOURCE node);
// every test that builds an ad hoc tree fixture should call this rather than redeclaring
// its own local mk(rel, body) closure.
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

// writeSpecTree explodes a monolithic topoSpec-shaped JSON document (the same shape the
// deleted monolithic topology.json form used: {"nodes":[...],"edges":[...]}) into the
// directory tree LoadTopology now requires, so a test can keep declaring its fixture as
// one concise literal while exercising the real on-disk input form (per
// memory/feedback_headless_repro_verifies_persistence.md).
//
// It writes, per node: nodes/<id>/meta.json (every node-level key the spec carried,
// minus data/inputs/outputs), nodes/<id>/data.json (when data is present),
// nodes/<id>/inputs/<name>.json and nodes/<id>/outputs/<name>.json per port; per edge:
// nodes/<source>/edges/<label>.json (adjacency layout — the redundant "source" key is
// dropped, since loadTree derives it from the containing node directory).
//
// Returns root so callers can pass it straight to LoadTopology.
func writeSpecTree(t *testing.T, root string, specJSON string) string {
	t.Helper()
	var spec topoSpec
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		t.Fatalf("writeSpecTree: parse spec JSON: %v", err)
	}

	for _, n := range spec.Nodes {
		meta := jsonMeta{
			ID:              n.ID,
			Type:            n.Type,
			R:               n.R,
			ScenePolarR:     n.ScenePolarR,
			ScenePolarTheta: n.ScenePolarTheta,
			ScenePolarPhi:   n.ScenePolarPhi,
			QuantITheta:     n.QuantITheta,
			QuantIPhi:       n.QuantIPhi,
			QuantIR:         n.QuantIR,
			StepTheta:       n.StepTheta,
			StepPhi:         n.StepPhi,
			StepR:           n.StepR,
			Gate:            n.Gate,
		}
		metaBody, err := json.Marshal(meta)
		if err != nil {
			t.Fatalf("writeSpecTree: marshal meta for %q: %v", n.ID, err)
		}
		writeTreeFile(t, root, filepath.Join("nodes", n.ID, "meta.json"), string(metaBody))

		if n.Data != nil {
			dataBody, err := json.Marshal(n.Data)
			if err != nil {
				t.Fatalf("writeSpecTree: marshal data for %q: %v", n.ID, err)
			}
			writeTreeFile(t, root, filepath.Join("nodes", n.ID, "data.json"), string(dataBody))
		}
	}

	for _, e := range spec.Edges {
		src := e.Source
		e.Source = "" // adjacency layout: the source is the directory, not a field on disk
		body, err := json.Marshal(e)
		if err != nil {
			t.Fatalf("writeSpecTree: marshal edge %q: %v", e.Label, err)
		}
		writeTreeFile(t, root, filepath.Join("nodes", src, "edges", e.Label+".json"), string(body))
	}

	return root
}

// WriteSpecTree is writeSpecTree, exported so external test packages (package
// Wiring_test files, e.g. speed_delivery_full_set_test.go) can reuse the same fixture
// builder instead of redeclaring it. This is a _test.go-only export — it never ships in
// production builds (go test excludes _test.go files from `go build`/library archives),
// so it does not violate "no test helper exported into production code."
func WriteSpecTree(t *testing.T, root string, specJSON string) string {
	t.Helper()
	return writeSpecTree(t, root, specJSON)
}

// quantizedDragTarget returns the position a drag to target actually COMMITS to under the
// scene lattice — bead CRUD (bead_crud.go, PLAN.md "moving a node is CRUD on the edge
// beads touching it"): every touching bead (dragTouchingBeads) judges the SAME raw target
// independently, and resolveBeadCrudMove resolves those verdicts into the node's single
// committed centre — taken from the BEAD OPERATION along that edge's own chain axis, NEVER
// from the raw target itself (the raw target is used directly only for a FREE node with no
// touching beads at all) — the raw target unchanged when quantizedLayout is off. MUST be
// called BEFORE the drag commits (reads the node's pre-drag center and its neighbours'
// pre-drag centers as the judging configuration) — calls the EXACT SAME
// resolveBeadCrudMove commitNodeMoveLocal calls, so this is not an independent oracle of
// the formula, only of the call.
func quantizedDragTarget(md *MoveDispatch, nodeID string, target vec3) vec3 {
	if !md.lq.quantizedLayout {
		return target
	}
	nm, ok := md.mr.nodeGeoms[nodeID]
	if !ok {
		return target
	}
	prev, ok := md.centerOfNode(nodeID)
	if !ok {
		return target
	}
	beads := dragTouchingBeads(md, nm, prev)
	if len(beads) == 0 {
		return target
	}
	committed, _ := resolveBeadCrudMove(beads, prev, target, wire.BeadStepR)
	return committed
}
