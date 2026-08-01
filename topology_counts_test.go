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
	"os"
	"path/filepath"
	"testing"
)

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
