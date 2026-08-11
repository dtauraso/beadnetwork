// scene_tabs_test.go — the pure nodes/Wiring/scene helpers (SelectedSceneIndex,
// AnchorIsTabbed, ResolveScenePath, SceneTabNames) verified through real bytes on disk.
// SelectScene itself, and the tests that drove it, moved to
// nodes/Wiring/sceneswitch/select_scene_test.go alongside SelectScene
// (docs/planning/movedispatch-decomposition.md, the remainder cluster) — those tests never
// needed *MoveDispatch, only *sceneswitch.SceneSwitch.
package scene_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
)

// tabbedAnchor builds a temp anchor whose basename is tab 0's Dir, since that is what makes
// an anchor tabbed at all (AnchorIsTabbed).
func tabbedAnchor(t *testing.T) string {
	t.Helper()
	anchor := filepath.Join(t.TempDir(), scene.SceneTabs[0].Dir)
	if err := os.MkdirAll(anchor, 0o755); err != nil {
		t.Fatalf("mkdir anchor: %v", err)
	}
	return anchor
}

func TestSelectedSceneIndexFallsBackToTabZero(t *testing.T) {
	anchor := tabbedAnchor(t)
	viewDir := filepath.Join(anchor, "view")
	if err := os.MkdirAll(viewDir, 0o755); err != nil {
		t.Fatalf("mkdir view: %v", err)
	}
	for _, tc := range []struct{ name, body string }{
		{"malformed", "{not json"},
		{"unknown tab name", `{"selected":"a-scene-that-was-deleted"}`},
		{"empty", `{}`},
	} {
		if err := os.WriteFile(filepath.Join(viewDir, "scene.json"), []byte(tc.body), 0o644); err != nil {
			t.Fatalf("write scene.json: %v", err)
		}
		if got := scene.SelectedSceneIndex(anchor); got != 0 {
			t.Fatalf("%s scene.json: SelectedSceneIndex = %d, want 0 — an unreadable selection must load the default scene, not fail to start", tc.name, got)
		}
	}
}

// An UNTABBED anchor (every test fixture, every one-off tree) must load exactly itself and
// stream no tabs — that is what keeps the strip absent instead of showing one dead tab, and
// what lets every existing fixture keep working without growing a scenes layout.
func TestUntabbedAnchorLoadsItselfAndHasNoTabs(t *testing.T) {
	anchor := filepath.Join(t.TempDir(), "some-fixture")
	if err := os.MkdirAll(anchor, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if scene.AnchorIsTabbed(anchor) {
		t.Fatalf("AnchorIsTabbed(%q) = true; only a tree named %q is the tabbed anchor", anchor, scene.SceneTabs[0].Dir)
	}
	if got := scene.ResolveScenePath(anchor); got != anchor {
		t.Fatalf("ResolveScenePath(%q) = %q, want the anchor itself", anchor, got)
	}
	if got := scene.SceneTabNames(anchor); len(got) != 0 {
		t.Fatalf("SceneTabNames(%q) = %v, want none", anchor, got)
	}
}

// A tabbed anchor resolves each tab to a SIBLING directory, never a child — the scenes are
// peers, and the anchor is one of them.
func TestResolveScenePathPicksTheSelectedSibling(t *testing.T) {
	anchor := tabbedAnchor(t)
	parent := filepath.Dir(anchor)

	if got, want := scene.ResolveScenePath(anchor), anchor; got != want {
		t.Fatalf("with no selection, ResolveScenePath = %q, want the anchor %q", got, want)
	}
	if err := scenepersist.WriteSelectedScene(anchor, 1); err != nil {
		t.Fatalf("WriteSelectedScene: %v", err)
	}
	if got, want := scene.ResolveScenePath(anchor), filepath.Join(parent, scene.SceneTabs[1].Dir); got != want {
		t.Fatalf("ResolveScenePath = %q, want %q", got, want)
	}
	if got, want := len(scene.SceneTabNames(anchor)), len(scene.SceneTabs); got != want {
		t.Fatalf("SceneTabNames returned %d labels, want %d — the strip must show every tab, not only the selected one", got, want)
	}
}
