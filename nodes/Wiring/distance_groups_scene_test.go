// distance_groups_scene_test.go — the three named distance groups belong to the RING, and
// only the ring resolves them.
//
// This is the case the flag exists for: node ids are per-scene directory names, so the
// ring's "input" group (which holds the pair (1, 2)) also NAMES two real nodes in the pair
// scene, whose nodes are likewise called 1 and 2. Before SceneTab.DistanceGroups, that group
// resolved against them and the distance panel appeared in the pair tab showing a live
// length for a group that was never about those nodes.
//
// Both assertions are needed and they fail in opposite directions: without the "ring still
// resolves" case a flag stuck at false would look correct, and without the "pair resolves
// nothing" case the bug itself comes back. Each test drives ONE MoveDispatch, reads what
// that one goroutine computed, and starts no mover network (docs/testing-shape.md).
package Wiring

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// loadSceneMD loads a scene tree by directory name and resolves its distance-group flag the
// way main.go does at startup.
func loadSceneMD(t *testing.T, sceneDir string) *MoveDispatch {
	t.Helper()
	root := filepath.Join(repoRootForDistanceGroupsTest(t), sceneDir)
	tr := T.NewWithSinkHook(nil, nil)
	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology(%s): %v", sceneDir, err)
	}
	md.ResolveSceneDistanceGroups(root)
	return md
}

// TestRingResolvesItsDistanceGroups is the guard against the flag being stuck off: the ring
// owns these groups, so at least one of the three must come back with a length.
func TestRingResolvesItsDistanceGroups(t *testing.T) {
	md := loadSceneMD(t, "topology")
	timeLen, inputLen, gateLen := md.DistanceGroupLens()
	if timeLen == 0 && inputLen == 0 && gateLen == 0 {
		t.Fatal("ring streamed all three group lengths as 0 — the ring owns these groups, so the panel would not render")
	}
}

// TestPairSceneIsDeniedTheGroups is the guard against the original defect. It cannot LOAD
// the pair tree the way the ring test does: the pair's kinds (Node1/Node2) import Wiring, so
// this package's own test binary cannot import them back to register them, and the load
// fails with `unknown type "Node1"`. So it asserts the two halves of the defect separately —
//
//	(a) the flag says no for the pair scene, and
//	(b) the pair tree really does contain node dirs 1 and 2,
//
// which together are the claim: the ids the ring's "input" group names DO exist over there,
// and the flag is what stops that group being read against them. Without (b) this test would
// keep passing if the pair scene were renumbered, having quietly stopped covering anything.
func TestPairSceneIsDeniedTheGroups(t *testing.T) {
	root := repoRootForDistanceGroupsTest(t)
	pair := filepath.Join(root, "topology-pair")
	if SceneHasDistanceGroups(pair) {
		t.Fatal("SceneHasDistanceGroups(topology-pair) = true, want false — the ring's groups are not the pair's")
	}
	if !SceneHasDistanceGroups(filepath.Join(root, "topology")) {
		t.Fatal("SceneHasDistanceGroups(topology) = false, want true — the ring owns these groups")
	}
	if SceneHasDistanceGroups(filepath.Join(root, "no-such-scene")) {
		t.Fatal("an unknown tree claimed the ring's distance groups, want false")
	}
	for _, id := range []string{"1", "2"} {
		if _, err := os.Stat(filepath.Join(pair, "nodes", id)); err != nil {
			t.Fatalf("pair scene has no node dir %q (%v) — this test no longer covers the id collision it was written for", id, err)
		}
	}
	// The ring's "input" group is the one that collided; if it stops naming a pair id, the
	// collision this guards is gone and the test should be rewritten rather than left to pass.
	named := false
	for _, p := range distanceGroups["input"] {
		if p.Source == "1" || p.Target == "1" || p.Source == "2" || p.Target == "2" {
			named = true
		}
	}
	if !named {
		t.Fatal("the ring's \"input\" group no longer names node 1 or 2 — the collision this test covers no longer exists")
	}
}

// TestGroupsAreInertUntilResolved: hasDistanceGroups defaults to FALSE on a freshly built
// MoveDispatch, so a scene that never calls ResolveSceneDistanceGroups resolves nothing —
// including ApplyDistanceGroupTarget, which reads the same distanceGroupMax the lengths do
// and so must refuse rather than move a node to satisfy a group the scene may not have.
func TestGroupsAreInertUntilResolved(t *testing.T) {
	root := filepath.Join(repoRootForDistanceGroupsTest(t), "topology")
	tr := T.NewWithSinkHook(nil, nil)
	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology(topology): %v", err)
	}
	// Deliberately NOT calling md.ResolveSceneDistanceGroups.
	if timeLen, inputLen, gateLen := md.DistanceGroupLens(); timeLen != 0 || inputLen != 0 || gateLen != 0 {
		t.Fatalf("unresolved dispatch streamed (%v, %v, %v), want all 0", timeLen, inputLen, gateLen)
	}
	for i := range distanceGroupOrder {
		if ok := md.ApplyDistanceGroupTarget(i, 1); ok {
			t.Fatalf("ApplyDistanceGroupTarget(%d, up) = true before the scene was resolved, want false", i)
		}
	}
}
