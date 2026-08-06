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
	// QuantizedDrag selects which DRAG this scene uses, per scene, because the two are
	// genuinely different behaviours rather than a tuning knob:
	//
	//   true  — the node is drawn from its QUANTIZED polar triple, so it steps one bead
	//           distance (wire.BeadStepR) at a time, exactly like the beads on its own
	//           chains. Commit 0a60ffb6 made this the behaviour, fixing the complaint that
	//           "when I move a node it's jump is very very small. when a bead moves it's
	//           jump is multiple times larger" — the node used to glide continuously while
	//           its beads jumped, because the raw target was drawn and the quantized one
	//           persisted.
	//   false — the pre-0a60ffb6 drag: the node follows the pointer continuously and no
	//           offset is measured (quantized_move.go's commitNodeMoveLocal already carries
	//           this branch; this field is what makes it reachable per scene).
	//
	// A step is only invisible when it is small against the scene. The ring spans ~500
	// world units, so a ~9-unit step reads as smooth; a two-node scene 40 units across
	// moves ~22% of itself per step, which is why the pair could not be dragged at all.
	//
	// NO COMMITTED SCENE USES THE QUANTIZED DRAG TODAY. The pair never did, and the ring
	// moved off it deliberately: a node is ONE POINT, an edge is the distance between two
	// of them, and a drag says where the point went — the bead count then fills whatever
	// line that leaves (edgeStepCount, from the live centre-to-centre distance, which is
	// already how the count is derived and did not change). Under the quantized drag the
	// node's new centre came from the bead operation along the chain axis instead, so the
	// drag target only nominated a direction.
	//
	// It remains the default for an unrecognised tree, so the bead-CRUD path is still
	// reachable and still tested — but nothing a user opens exercises it. If that stays
	// true, the honest next step is to delete that path rather than leave a large unused
	// mechanism reading as live; that is a model decision, not a cleanup, so it is named
	// here rather than taken.
	QuantizedDrag bool
	// CoplanarEdges makes a node's RING PLANE contain the edge leaving it, so the bead
	// chain, the node's torus and the beads' own tori all lie in ONE plane instead of the
	// chain running through the tori's holes.
	//
	// The ring's axis is normally the node's INWARD pole (toward the scene centre), which
	// says nothing about where its neighbour is — so an edge lies in that plane only by
	// coincidence. With this on, the pole is projected PERPENDICULAR to the edge: the
	// closest axis to the inward one whose plane still contains the edge.
	//
	// Only meaningful for a node with ONE neighbour. No single plane contains two edges
	// that are not collinear, so a node with more keeps its inward pole and this is inert
	// — which is why it is a per-scene choice rather than a global rule.
	CoplanarEdges bool
	// UpAxis aims this scene's node tori — and the per-node vector drawn along the same
	// axis — at world +y, straight up, instead of at anything derived from where the node
	// sits. For a scene whose nodes share a height (the pair does: both at y=0) an up axis
	// ALSO contains the edge between them, so it satisfies CoplanarEdges at the same time;
	// the two are separate fields because that agreement is a property of this scene's
	// layout, not a general fact.
	UpAxis bool
	// ClockDivisor slows this scene's EFFECTIVE clock speed relative to the user's chosen
	// multiplier: effective = userSpeed / ClockDivisor. It is a property of the SCENE, not a
	// tuning knob on the user's speed — the user's number in the slider (and in
	// view/speed.json) is unchanged; only the rate actually reaching the clocks is scaled.
	//
	// The ring spans ~500 world units, so one bead's step across it reads as a small
	// fraction of the scene. The pair is ~40 units across between its two nodes, so the
	// SAME wall-clock pace moves a bead a much larger fraction of the scene per tick — the
	// same user-facing "speed" setting reads far faster on the pair than on the ring purely
	// because the pair is smaller, not because the user asked for anything different.
	// ClockDivisor corrects for that so both scenes read at a comparable felt pace.
	//
	// Ring = 1 (no correction needed). Pair = 4 (runs at 1/4 the ring's clock for the same
	// slider setting). Never 0 or negative — EffectiveClockSpeed guards against a
	// division by an invalid value from an unrecognised or malformed scene.
	ClockDivisor float64
}

// SceneTabs is the tab strip, in display order. Index 0 is the DEFAULT: its Dir must be
// the anchor's own basename, since that is the path the extension host launches with and
// sizes its stream fds from (see AnchorIsTabbed).
var SceneTabs = []SceneTab{
	{Name: "ring", Dir: "topology", QuantizedDrag: false, CoplanarEdges: false, UpAxis: false, ClockDivisor: 1},
	{Name: "pair", Dir: "topology-pair", QuantizedDrag: false, CoplanarEdges: true, UpAxis: true, ClockDivisor: 4},
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

// SceneUsesQuantizedDrag answers, for the tree actually being LOADED, whether the node
// drag snaps to the bead lattice. It takes the loaded scene's own path (not the anchor)
// because the loader knows which tree it is opening but not which tab pointed it there.
// An unknown tree — every test fixture, every one-off run — gets the quantized drag, which
// is what every scene did before scenes were selectable.
func SceneUsesQuantizedDrag(scenePath string) bool {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.QuantizedDrag
		}
	}
	return true
}

// SceneWantsCoplanarEdges answers, for the tree being LOADED, whether a node's ring plane
// must contain the edge leaving it (SceneTab.CoplanarEdges). Unknown trees keep the plain
// inward pole, which is what every scene had before this was a choice.
func SceneWantsCoplanarEdges(scenePath string) bool {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.CoplanarEdges
		}
	}
	return false
}

// SceneWantsUpAxis answers whether the tree being LOADED aims its node tori and per-node
// vectors straight up (SceneTab.UpAxis). Unknown trees do not — they keep the unrotated
// ring every scene had before ring orientation existed.
func SceneWantsUpAxis(scenePath string) bool {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.UpAxis
		}
	}
	return false
}

// SceneClockDivisor answers, for the tree being LOADED, its SceneTab.ClockDivisor. A test
// fixture or one-off tree with no tab entry gets divisor 1 (no scaling) — never 0, which
// would divide the effective speed by zero downstream.
func SceneClockDivisor(scenePath string) float64 {
	base := filepath.Base(filepath.Clean(scenePath))
	for _, t := range SceneTabs {
		if t.Dir == base {
			return t.ClockDivisor
		}
	}
	return 1
}
