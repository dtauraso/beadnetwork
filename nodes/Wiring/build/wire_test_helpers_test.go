package build_test

// wire_test_helpers_test.go — fixture-tree writers for this package's LoadTopology tests.
// Duplicated from nodes/Wiring/dispatch's own writeTreeFile/writeSpecTree (a _test.go-only
// export, usable only by test files in THAT package's own directory — the trick does not
// cross directories, so a genuine copy is needed here) rather than sharing across packages
// (docs/planning/movedispatch-decomposition.md §34).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
)

// writeTreeFile writes body to <root>/<rel>, creating any missing parent directories.
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
// It writes, per node: nodes/<id>/meta.json (every node-level key the spec carried, minus
// data/inputs/outputs), nodes/<id>/data.json (when data is present); per edge:
// nodes/<source>/edges/<label>.json (adjacency layout — the redundant "source" key is
// dropped, since loadTree derives it from the containing node directory).
//
// Returns root so callers can pass it straight to LoadTopology.
func writeSpecTree(t *testing.T, root string, specJSON string) string {
	t.Helper()
	var spec loadspec.TopoSpec
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		t.Fatalf("writeSpecTree: parse spec JSON: %v", err)
	}

	for _, n := range spec.Nodes {
		meta := loadspec.JSONMeta{
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
