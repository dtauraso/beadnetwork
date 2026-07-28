package Wiring

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

// writeTreeFile writes body to <root>/<rel>, creating any missing parent directories.
// It is the package's single fixture-file writer for directory-tree topology test
// fixtures (nodes/<id>/meta.json, nodes/<id>/inputs|outputs/<name>.json,
// edges/<label>.json); every test that builds an ad hoc tree fixture should call this
// rather than redeclaring its own local mk(rel, body) closure.
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
// list of (source, target) pairs the fixture wrote under edges/. Cascade adjacency here
// is the full domain adjacency: fixtures exercising the DRAG/re-quantize fan want every
// domain neighbor reachable, and a fixture that needs a narrower cascade set (to exercise
// a per-kind relay rule) should write its own files instead of calling this.
func writeCascadeEdgesFromEdges(t *testing.T, root string, nodeKinds map[string]string, edges [][2]string) {
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
		if len(n) == 0 {
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

// cascadeSettle is the fixed wall-clock window some drag/neighbor tests sleep after a
// polled convergence, to let any (unwanted) further cascade land before asserting
// absence. It is a widening window, NOT the proof of absence — the proof is an
// "abc-drag" breadcrumb count check (see neighbor_setc_test.go / drag_persist_e2e_test.go,
// abcDragDeltasFor): a fixed sleep alone can silently pass for the wrong reason under
// load.
const cascadeSettle = 20 * time.Millisecond

// approxEq is the float tolerance used by geometry/position wire tests.
func approxEq(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
