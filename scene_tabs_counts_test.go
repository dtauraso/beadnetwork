// scene_tabs_counts_test.go — the two facts that make a SECOND scene safe to select from
// the tab strip. Both are properties of the committed trees, checked by walking them
// directly (same approach and reasoning as topology_counts_test.go, whose helpers this
// reuses).
//
// Every tab is a real, complete topology tree, so each carries its own hand-maintained
// counts.json — and the extension host reads only the ANCHOR'S, because it must size the
// child's stdio array before Go exists to be asked (.claude/rules/persistence-ownership.md,
// "Counts are stored, never re-derived"). That is the trap this file exists to catch: a
// scene needing MORE rows than the anchor's counts allocate would silently lose the streams
// past the end, drawing an incomplete diagram with no error anywhere.
package main

import (
	"os"
	"path/filepath"
	"testing"

	W "github.com/dtauraso/wirefold/nodes/Wiring"
)

// sceneRootFor resolves a tab's committed tree from the repo root. It mirrors
// ResolveScenePath's own rule (a tab Dir is a SIBLING of the anchor) without calling it,
// since ResolveScenePath answers "which one is selected" and this test asks about all of
// them regardless of selection.
func sceneRootFor(repoRoot string, tab W.SceneTab) string {
	return filepath.Join(repoRoot, tab.Dir)
}

// TestEverySceneTabCountsMatchItsTree — each tab's counts.json describes its OWN tree.
// Without this, only the anchor's counts were ever checked, and a second scene's file could
// claim anything.
func TestEverySceneTabCountsMatchItsTree(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	for _, tab := range W.SceneTabs {
		root := sceneRootFor(repoRoot, tab)
		stored := topologyCounts(t, root)
		wantNodes := len(nodeRowLabels(t, root))
		wantEdges := countEdgeFiles(t, root)
		if stored.Nodes != wantNodes || stored.Edges != wantEdges {
			t.Fatalf("scene %q (%s): counts.json claims {nodes:%d edges:%d} but the tree has "+
				"{nodes:%d edges:%d} — counts.json is hand-maintained; edit it to match the tree",
				tab.Name, tab.Dir, stored.Nodes, stored.Edges, wantNodes, wantEdges)
		}
	}
}

// TestNoSceneTabExceedsTheAnchorsCounts — the anchor's counts.json must cover every scene,
// because it is the ONLY one the extension host reads (it sizes one dedicated fd per node
// row and per edge before spawning Go, and the anchor path never changes when a tab does).
// A scene with more rows than that would have its extra per-node/per-edge streams land on
// fds nobody allocated: no error, just missing geometry.
//
// The bound is one-directional on purpose. FEWER rows than the anchor is fine — the surplus
// fds simply carry no frames — so this asserts a ceiling, not equality.
func TestNoSceneTabExceedsTheAnchorsCounts(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	anchor := topologyCounts(t, sceneRootFor(repoRoot, W.SceneTabs[0]))
	for _, tab := range W.SceneTabs[1:] {
		scene := topologyCounts(t, sceneRootFor(repoRoot, tab))
		if scene.Nodes > anchor.Nodes || scene.Edges > anchor.Edges {
			t.Fatalf("scene %q (%s) needs {nodes:%d edges:%d} but the ANCHOR %q allocates only "+
				"{nodes:%d edges:%d}; the extension host sizes its stdio array from the anchor's "+
				"counts.json alone, so this scene's extra streams would land on unallocated fds "+
				"and its geometry would silently not draw. Raise the anchor's counts.json, or "+
				"shrink this scene",
				tab.Name, tab.Dir, scene.Nodes, scene.Edges, W.SceneTabs[0].Dir, anchor.Nodes, anchor.Edges)
		}
	}
}

// countEdgeFiles counts every nodes/<id>/edges/<label>.json in a scene tree — the same
// "edges" definition topology/counts.json stores (edges have no id space to gap, so it is a
// plain count, not a row count).
func countEdgeFiles(t *testing.T, sceneRoot string) int {
	t.Helper()
	nodeEntries, err := os.ReadDir(filepath.Join(sceneRoot, "nodes"))
	if err != nil {
		t.Fatalf("ReadDir %s/nodes: %v", sceneRoot, err)
	}
	n := 0
	for _, ne := range nodeEntries {
		if !ne.IsDir() {
			continue
		}
		edgeEntries, err := os.ReadDir(filepath.Join(sceneRoot, "nodes", ne.Name(), "edges"))
		if err != nil {
			if os.IsNotExist(err) {
				continue // a node with no outgoing edges has no edges/ dir at all
			}
			t.Fatalf("ReadDir %s/nodes/%s/edges: %v", sceneRoot, ne.Name(), err)
		}
		for _, ee := range edgeEntries {
			if !ee.IsDir() {
				n++
			}
		}
	}
	return n
}
