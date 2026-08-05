// scene_tabs.go — the SCENE TABS: which diagrams this editor can show, which one is
// showing, and how a click on a tab becomes the other diagram.
//
// OWNERSHIP: Go owns all three. The tab list is this file's SceneTabs; the selection is
// Go's own persisted state (<anchor>/view/scene.json); the switch is performed by Go. TS
// renders the strip from the VIEW frame and forwards a click as one addressed edit
// (edit-update kind="scene" attr="selected"). It holds no list, no labels, no selection —
// same shape as the overlay toggles (see overlay_gen.go / overlay-flags.ts).
//
// THE ANCHOR vs THE SCENE. The -topology flag is the ANCHOR: the fixed path the extension
// host launches against and reads counts.json from. It never changes for the life of the
// editor. The SCENE is the directory actually loaded, resolved from the anchor's parent by
// the selected tab's Dir. Keeping selection state at the anchor (never inside the scene it
// selects) is what makes the selection readable before any scene has been chosen — a
// selection stored inside scene B would be unreachable while scene A is loaded.
//
// HOW THE SWITCH HAPPENS — no in-process teardown, and no new TS restart path. SelectScene
// persists the new selection and asks the process to end. runCommand.ts's runner is already
// LOOPING (runCommand.ts sets looping = true on every successful run), so a natural exit is
// already followed by a respawn against the same anchor; that respawn re-reads this file's
// selection and loads the other scene. Killing live node goroutines and their in-flight
// beads mid-traversal — an in-process rebuild — buys nothing over the respawn that the .go
// file watcher already performs on every Go edit.
package Wiring

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SceneTab is one tab: the label Go streams to the strip, and the directory it loads.
// Dir is resolved relative to the ANCHOR'S PARENT, so the scenes are sibling topology
// trees — each one a complete, independently loadable directory tree with its own
// counts.json and its own view/ state.
type SceneTab struct {
	Name string
	Dir  string
}

// SceneTabs is the tab strip, in display order. Index 0 is the DEFAULT: its Dir must be
// the anchor's own basename, since that is the path the extension host launches with and
// sizes its stream fds from (see AnchorIsTabbed).
var SceneTabs = []SceneTab{
	{Name: "ring", Dir: "topology"},
	{Name: "pair", Dir: "topology-pair"},
}

// sceneSelectionFile is the persisted selection, held at the ANCHOR (never inside a scene).
type sceneSelectionFile struct {
	Selected string `json:"selected"`
}

// AnchorIsTabbed reports whether this anchor is the tabbed layout's anchor — i.e. whether
// its basename is tab 0's Dir. A test fixture or a one-off tree launched from somewhere
// else is NOT tabbed: it loads exactly itself, streams an empty tab strip, and no tab UI
// appears. That keeps every existing fixture working untouched rather than requiring each
// to grow a scenes layout.
func AnchorIsTabbed(anchorPath string) bool {
	return filepath.Base(filepath.Clean(anchorPath)) == SceneTabs[0].Dir
}

// SelectedSceneIndex reads the persisted selection for this anchor. An absent, malformed,
// or unknown-name file means tab 0 — a fresh checkout has no selection, and that is the
// normal case, not an error.
func SelectedSceneIndex(anchorPath string) int {
	if !AnchorIsTabbed(anchorPath) {
		return 0
	}
	b, err := os.ReadFile(sceneSelectionFilePath(anchorPath))
	if err != nil {
		return 0
	}
	var f sceneSelectionFile
	if json.Unmarshal(b, &f) != nil {
		return 0
	}
	for i, t := range SceneTabs {
		if t.Name == f.Selected {
			return i
		}
	}
	return 0
}

// ResolveScenePath maps an anchor to the directory LoadTopology should actually load: the
// selected tab's sibling directory, or the anchor itself when this anchor is not tabbed.
func ResolveScenePath(anchorPath string) string {
	if !AnchorIsTabbed(anchorPath) {
		return anchorPath
	}
	tab := SceneTabs[SelectedSceneIndex(anchorPath)]
	return filepath.Join(filepath.Dir(filepath.Clean(anchorPath)), tab.Dir)
}

// SceneTabNames is the label list streamed on the VIEW frame. Empty for an untabbed
// anchor, which is what makes the strip absent rather than a single dead tab.
func SceneTabNames(anchorPath string) []string {
	if !AnchorIsTabbed(anchorPath) {
		return nil
	}
	names := make([]string, len(SceneTabs))
	for i, t := range SceneTabs {
		names[i] = t.Name
	}
	return names
}

// sceneSwitch is MoveDispatch's half: the anchor to persist against, and the way to end
// this process so the runner's looping respawn loads the newly selected scene. Both nil/
// empty until EnableSceneSwitch arms them, so a bare test-constructed MoveDispatch cannot
// exit anything.
type sceneSwitch struct {
	anchorPath string
	quit       func()
}

// EnableSceneSwitch arms tab switching. quit ends the run (main's context cancel), which
// the extension host's looping runner follows with a respawn. Called from main.go after
// load, before the stdin reader starts — the view-owner goroutine is the only caller of
// SelectScene, so this field is written once, before that goroutine exists.
func (md *MoveDispatch) EnableSceneSwitch(anchorPath string, quit func()) {
	md.scenes.anchorPath = anchorPath
	md.scenes.quit = quit
}

// SelectScene handles a tab click: persist, then end the run so the respawn loads it.
// Selecting the tab already showing is a no-op — restarting the sim to arrive at the same
// diagram would look like a random flicker to whoever clicked.
func (md *MoveDispatch) SelectScene(idx int) {
	if md.scenes.anchorPath == "" || md.scenes.quit == nil {
		return
	}
	if idx < 0 || idx >= len(SceneTabs) {
		return
	}
	if idx == SelectedSceneIndex(md.scenes.anchorPath) {
		return
	}
	if err := writeSelectedScene(md.scenes.anchorPath, idx); err != nil {
		// Do NOT quit on a failed write: the respawn would reload the OLD scene and the
		// click would read as "the editor restarted for no reason". Report and stay put.
		// stderr, not a breadcrumb: the extension host pipes this straight to the sim's
		// output channel AND .probe/go-errors.jsonl, which is where an operator looks when
		// a click did nothing (memory/feedback_runner_errors_probe_first.md).
		fmt.Fprintf(os.Stderr, "scene tab: could not persist selection to %s: %v — staying on the current scene\n",
			sceneSelectionFilePath(md.scenes.anchorPath), err)
		return
	}
	md.scenes.quit()
}
