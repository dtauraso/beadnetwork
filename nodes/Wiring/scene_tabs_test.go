// scene_tabs_test.go — what ONE goroutine (the view-owner) decides and persists when a tab
// is clicked, verified through real bytes on disk (docs/testing-shape.md's one exception:
// persistence through a real write/read round trip).
//
// Nothing here runs two goroutines: SelectScene's quit func is a plain recorder, so these
// assert the decision and the file, never a handoff.
package Wiring

import (
	"os"
	"path/filepath"
	"testing"

	T "github.com/dtauraso/wirefold/Trace"
)

// tabbedAnchor builds a temp anchor whose basename is tab 0's Dir, since that is what makes
// an anchor tabbed at all (AnchorIsTabbed).
func tabbedAnchor(t *testing.T) string {
	t.Helper()
	anchor := filepath.Join(t.TempDir(), SceneTabs[0].Dir)
	if err := os.MkdirAll(anchor, 0o755); err != nil {
		t.Fatalf("mkdir anchor: %v", err)
	}
	return anchor
}

// armedDispatch returns a MoveDispatch with scene switching armed, plus a pointer to the
// quit-called flag.
func armedDispatch(t *testing.T, anchor string) (*MoveDispatch, *bool) {
	t.Helper()
	md := &MoveDispatch{tr: T.New()}
	quit := false
	md.EnableSceneSwitch(anchor, func() { quit = true })
	return md, &quit
}

func TestSelectSceneWritesTheSelectionAndEndsTheRun(t *testing.T) {
	anchor := tabbedAnchor(t)
	md, quit := armedDispatch(t, anchor)

	md.SelectScene(1)

	if !*quit {
		t.Fatalf("SelectScene(1) did not end the run; without that the respawn never happens and the tab never changes")
	}
	// Read the selection back the way a fresh process would — not by inspecting a field.
	if got := SelectedSceneIndex(anchor); got != 1 {
		t.Fatalf("after SelectScene(1), a fresh read of the persisted selection = %d, want 1", got)
	}
	// And the bytes name the TAB, not its index (see writeSelectedScene's doc comment).
	raw, err := os.ReadFile(filepath.Join(anchor, "view", "scene.json"))
	if err != nil {
		t.Fatalf("read scene.json: %v", err)
	}
	if want := `{"selected":"` + SceneTabs[1].Name + `"}`; string(raw) != want {
		t.Fatalf("scene.json = %s, want %s", raw, want)
	}
}

func TestSelectSceneIgnoresTheTabAlreadyShowing(t *testing.T) {
	anchor := tabbedAnchor(t)
	md, quit := armedDispatch(t, anchor)

	md.SelectScene(0) // tab 0 is the default with no file present

	if *quit {
		t.Fatalf("selecting the tab already showing ended the run: the sim would restart and land on the SAME diagram, which reads as a random flicker")
	}
	if _, err := os.Stat(filepath.Join(anchor, "view", "scene.json")); !os.IsNotExist(err) {
		t.Fatalf("selecting the current tab wrote scene.json (err=%v); a no-op must not touch disk", err)
	}
}

func TestSelectSceneRejectsAnOutOfRangeTab(t *testing.T) {
	anchor := tabbedAnchor(t)
	md, quit := armedDispatch(t, anchor)

	md.SelectScene(len(SceneTabs)) // one past the end
	md.SelectScene(-1)

	if *quit {
		t.Fatalf("an out-of-range tab index ended the run; the respawn would reload the same scene for no reason")
	}
}

func TestUnarmedDispatchCannotEndTheRun(t *testing.T) {
	md := &MoveDispatch{tr: T.New()} // EnableSceneSwitch never called
	md.SelectScene(1)                // must not panic, must not exit
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
		if got := SelectedSceneIndex(anchor); got != 0 {
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
	if AnchorIsTabbed(anchor) {
		t.Fatalf("AnchorIsTabbed(%q) = true; only a tree named %q is the tabbed anchor", anchor, SceneTabs[0].Dir)
	}
	if got := ResolveScenePath(anchor); got != anchor {
		t.Fatalf("ResolveScenePath(%q) = %q, want the anchor itself", anchor, got)
	}
	if got := SceneTabNames(anchor); len(got) != 0 {
		t.Fatalf("SceneTabNames(%q) = %v, want none", anchor, got)
	}
}

// A tabbed anchor resolves each tab to a SIBLING directory, never a child — the scenes are
// peers, and the anchor is one of them.
func TestResolveScenePathPicksTheSelectedSibling(t *testing.T) {
	anchor := tabbedAnchor(t)
	parent := filepath.Dir(anchor)

	if got, want := ResolveScenePath(anchor), anchor; got != want {
		t.Fatalf("with no selection, ResolveScenePath = %q, want the anchor %q", got, want)
	}
	if err := writeSelectedScene(anchor, 1); err != nil {
		t.Fatalf("writeSelectedScene: %v", err)
	}
	if got, want := ResolveScenePath(anchor), filepath.Join(parent, SceneTabs[1].Dir); got != want {
		t.Fatalf("ResolveScenePath = %q, want %q", got, want)
	}
	if got, want := len(SceneTabNames(anchor)), len(SceneTabs); got != want {
		t.Fatalf("SceneTabNames returned %d labels, want %d — the strip must show every tab, not only the selected one", got, want)
	}
}
