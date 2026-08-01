// topology_counts_test.go — a fast, direct check that the hand-maintained
// topology/counts.json agrees with the real topology/ tree, without spawning the binary or
// the headless harness (memory/feedback_no_foreground_sim_runs.md; .claude/rules/persistence-ownership.md
// "Counts are stored, never re-derived" — no Go writer keeps this file in sync, so nothing
// else notices drift outside the headless spawn path, which costs seconds and only runs
// inside those tests). This test walks the tree directly and compares in milliseconds.
//
// "nodes" is the ROW COUNT (the largest node id under topology/nodes/, since row = id-1 —
// see nodeRowLabels's doc and headless_stream_helpers_test.go's spawnDedicatedAllStreams,
// which performs the same comparison as a byproduct of sizing its own spawn). "edges" is a
// plain count of every nodes/<id>/edges/<label>.json file in the tree (edges have no id
// space to gap — persistence-ownership.md).
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// nodeRowLabels returns, in ROW order (row = id-1 — loader_tree.go's loadTree parses each
// node directory name to an int and places it at that row directly; there is no ordering
// step left to recompute), the label each row's node is expected to stream. Under the
// current model this is simply the directory name itself: none of this repo's committed
// fixture nodes set data.label, so specNode.label()'s id fallback applies to every one.
// Recomputed from the actual directory listing (not hardcoded) so this test does not
// silently go stale if a node is added/removed from topology/. A gap row (no directory
// parses to that id) is left as "" — this repo's committed topology has no gaps, so every
// slot below RowCount is filled, but a caller must still treat "" as "no expectation for
// this row" rather than a real label.
//
// Moved here from the deleted headless_stream_helpers_test.go (task/row-fd-identity-parity):
// this is the last surviving caller now that the real-spawn headless tests are gone.
func nodeRowLabels(t *testing.T, repoRoot string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(repoRoot, "topology", "nodes"))
	if err != nil {
		t.Fatalf("ReadDir topology/nodes: %v", err)
	}
	ids := make([]int, 0, len(entries))
	rowCount := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		n, err := strconv.Atoi(e.Name())
		if err != nil {
			t.Fatalf("topology/nodes/%s: not a numeric node directory name: %v", e.Name(), err)
		}
		if n < 1 {
			t.Fatalf("topology/nodes/%s: node ids are 1-based, got %d", e.Name(), n)
		}
		ids = append(ids, n)
		if n > rowCount {
			rowCount = n
		}
	}
	rows := make([]string, rowCount)
	for _, n := range ids {
		rows[n-1] = strconv.Itoa(n)
	}
	return rows
}

type storedCounts struct {
	Nodes int `json:"nodes"`
	Edges int `json:"edges"`
}

// topologyCounts reads <repoRoot>/topology/counts.json. Fails the test loudly on a missing,
// malformed, or negative-count file rather than returning zero: a silent zero here would
// allocate no dedicated streams and the test would then "pass" against a bridge that never
// streamed anything (the exact degrade-invisibly behaviour readCounts was changed to stop).
//
// Moved here from the deleted headless_stream_helpers_test.go (task/row-fd-identity-parity):
// this is the last surviving caller now that the real-spawn headless tests are gone.
func topologyCounts(t *testing.T, repoRoot string) storedCounts {
	t.Helper()
	p := filepath.Join(repoRoot, "topology", "counts.json")
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	var c storedCounts
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatalf("parse %s: %v", p, err)
	}
	if c.Nodes < 0 || c.Edges < 0 {
		t.Fatalf("%s: counts must be non-negative, got %+v", p, c)
	}
	return c
}

func TestTopologyCountsMatchTree(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	stored := topologyCounts(t, repoRoot)

	wantNodes := len(nodeRowLabels(t, repoRoot))

	wantEdges := 0
	nodeEntries, err := os.ReadDir(filepath.Join(repoRoot, "topology", "nodes"))
	if err != nil {
		t.Fatalf("ReadDir topology/nodes: %v", err)
	}
	for _, ne := range nodeEntries {
		if !ne.IsDir() {
			continue
		}
		edgeEntries, err := os.ReadDir(filepath.Join(repoRoot, "topology", "nodes", ne.Name(), "edges"))
		if err != nil {
			if os.IsNotExist(err) {
				continue // a node with no outgoing edges has no edges/ dir at all
			}
			t.Fatalf("ReadDir topology/nodes/%s/edges: %v", ne.Name(), err)
		}
		for _, ee := range edgeEntries {
			if !ee.IsDir() {
				wantEdges++
			}
		}
	}

	if stored.Nodes != wantNodes || stored.Edges != wantEdges {
		t.Fatalf("topology/counts.json is stale: it claims {nodes:%d edges:%d} but the real "+
			"topology/ tree has {nodes:%d edges:%d} (nodes = largest id under topology/nodes/, "+
			"edges = count of nodes/*/edges/*.json files) — counts.json is hand-maintained "+
			"(.claude/rules/persistence-ownership.md, \"Counts are stored, never re-derived\"); "+
			"edit topology/counts.json to match the tree, do not change this test",
			stored.Nodes, stored.Edges, wantNodes, wantEdges)
	}
}
