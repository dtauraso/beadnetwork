package Wiring

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
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

// writeCascadeEdgesFromEdges writes a complete nodes/<id>/cascade-edges.json for EVERY
// node in the fixture, derived from the fixture's own edge list — cascade adjacency is
// mandatory (parseSpec's validateCascadeEdges) and a fixture that omits it fails to load.
//
// nodeKinds maps node id -> that node's Type (matching its meta.json), and edges is the
// list of (source, target) pairs the fixture wrote under each source node's edges/.
// Cascade adjacency here
// is the full domain adjacency: fixtures exercising the DRAG/re-quantize fan want every
// domain neighbor reachable, and a fixture that needs a narrower cascade set (to exercise
// a per-kind relay rule) should write its own files instead of calling this.
func writeCascadeEdgesFromEdges(t *testing.T, root string, nodeKinds map[string]string, edges [][2]string) {
	t.Helper()
	writeCascadeEdgesFromEdgesAllowIsolated(t, root, nodeKinds, edges, false)
}

// writeCascadeEdgesFromEdgesAllowIsolated is writeCascadeEdgesFromEdges's shared
// mechanism, parameterized on whether a node with zero domain neighbors is an error
// (requireNonEmpty=true, the original behavior every existing caller relies on to catch
// a fixture that forgot to wire a node in) or a legitimate isolated node that gets an
// empty cascade-edges.json (requireNonEmpty=false — validateCascadeEdges's equality rule
// permits this: no edges ⇒ no required cascade neighbors). writeSpecTree uses the
// latter, since a spec literal can legitimately declare a node with no edges.
func writeCascadeEdgesFromEdgesAllowIsolated(t *testing.T, root string, nodeKinds map[string]string, edges [][2]string, requireNonEmpty bool) {
	t.Helper()
	neighbors := map[string]map[string]string{}
	add := func(a, b string) {
		if _, ok := neighbors[a]; !ok {
			neighbors[a] = map[string]string{}
		}
		neighbors[a][b] = nodeKinds[b]
	}
	for _, e := range edges {
		add(e[0], e[1])
		add(e[1], e[0])
	}
	for id := range nodeKinds {
		n := neighbors[id]
		if requireNonEmpty && len(n) == 0 {
			t.Fatalf("fixture node %q has no domain neighbors; cascade adjacency is mandatory so every node needs at least one", id)
		}
		ids := make([]string, 0, len(n))
		for to := range n {
			ids = append(ids, to)
		}
		sort.Strings(ids)
		body, err := json.Marshal(struct {
			CascadeEdges []string          `json:"cascadeEdges"`
			CascadeKinds map[string]string `json:"cascadeKinds"`
		}{ids, n})
		if err != nil {
			t.Fatalf("marshal cascade-edges for %q: %v", id, err)
		}
		writeTreeFile(t, root, filepath.Join("nodes", id, "cascade-edges.json"), string(body))
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
// dropped, since loadTree derives it from the containing node directory); and, for every
// node, a nodes/<id>/cascade-edges.json derived from
// the spec's own edge list via writeCascadeEdgesFromEdgesAllowIsolated (cascade adjacency
// is mandatory per validateCascadeEdges — see that function's doc comment — and any
// cascadeEdges/cascadeKinds the spec literal itself carried are ignored in favor of this
// derivation, since domain adjacency is the only value that can satisfy the equality
// check anyway).
//
// Returns root so callers can pass it straight to LoadTopology.
func writeSpecTree(t *testing.T, root string, specJSON string) string {
	t.Helper()
	var spec topoSpec
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		t.Fatalf("writeSpecTree: parse spec JSON: %v", err)
	}

	nodeKinds := make(map[string]string, len(spec.Nodes))
	for _, n := range spec.Nodes {
		nodeKinds[n.ID] = n.Type

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
			LocalPolars:     n.LocalPolars,
			LocalPoleTheta:  n.LocalPoleTheta,
			LocalPolePhi:    n.LocalPolePhi,
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

		for _, p := range n.Inputs {
			body, err := json.Marshal(p)
			if err != nil {
				t.Fatalf("writeSpecTree: marshal input %q.%q: %v", n.ID, p.Name, err)
			}
			writeTreeFile(t, root, filepath.Join("nodes", n.ID, "inputs", p.Name+".json"), string(body))
		}
		for _, p := range n.Outputs {
			body, err := json.Marshal(p)
			if err != nil {
				t.Fatalf("writeSpecTree: marshal output %q.%q: %v", n.ID, p.Name, err)
			}
			writeTreeFile(t, root, filepath.Join("nodes", n.ID, "outputs", p.Name+".json"), string(body))
		}
	}

	edgePairs := make([][2]string, 0, len(spec.Edges))
	for _, e := range spec.Edges {
		edgePairs = append(edgePairs, [2]string{e.Source, e.Target})
		src := e.Source
		e.Source = "" // adjacency layout: the source is the directory, not a field on disk
		body, err := json.Marshal(e)
		if err != nil {
			t.Fatalf("writeSpecTree: marshal edge %q: %v", e.Label, err)
		}
		writeTreeFile(t, root, filepath.Join("nodes", src, "edges", e.Label+".json"), string(body))
	}

	writeCascadeEdgesFromEdgesAllowIsolated(t, root, nodeKinds, edgePairs, false)

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

// cascadeSettle is the fixed wall-clock window some drag/neighbor tests sleep after a
// polled convergence, to let any (unwanted) further cascade land before asserting
// absence. It is a widening window, NOT the proof of absence — the proof is an
// "abc-drag" breadcrumb count check (see neighbor_setc_test.go / drag_persist_e2e_test.go,
// abcDragDeltasFor): a fixed sleep alone can silently pass for the wrong reason under
// load.
const cascadeSettle = 20 * time.Millisecond
