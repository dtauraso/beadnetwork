// scene_tabs.go — the SCENE TAB REGISTRY: which diagrams this editor can show and what
// each one declares about itself.
//
// OWNERSHIP: Go owns all three (the tab list, the selection, and the switch — see this
// file's sibling scene_switch.go for the switch and scene_selection.go for anchor/path
// resolution against the selection). TS renders the strip from the VIEW frame and forwards
// a click as one addressed edit (edit-update kind="scene" attr="selected"). It holds no
// list, no labels, no selection — same shape as the overlay toggles (see overlay_gen.go /
// overlay-flags.ts).
//
// THE ANCHOR vs THE SCENE. The -topology flag is the ANCHOR: the fixed path the extension
// host launches against and reads counts.json from. It never changes for the life of the
// editor. The SCENE is the directory actually loaded, resolved from the anchor's parent by
// the selected tab's Dir. Keeping selection state at the anchor (never inside the scene it
// selects) is what makes the selection readable before any scene has been chosen — a
// selection stored inside scene B would be unreachable while scene A is loaded.
//
// This package holds the pure tab data and path resolution; the switch itself (the
// *MoveDispatch method SelectScene, plus runtopology/topology_run.go setting
// md.Scenes.AnchorPath/md.Scenes.Quit directly) lives in package Wiring's scene_switch.go,
// since a method on another package's type cannot live here.
//
// HOW THE SWITCH HAPPENS — no in-process teardown, and no new TS restart path. SelectScene
// persists the new selection and asks the process to end. runCommand.ts's runner is already
// LOOPING (runCommand.ts sets looping = true on every successful run), so a natural exit is
// already followed by a respawn against the same anchor; that respawn re-reads this file's
// selection and loads the other scene. Killing live node goroutines and their in-flight
// beads mid-traversal — an in-process rebuild — buys nothing over the respawn that the .go
// file watcher already performs on every Go edit.
package scene

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
	//           distance (lattice.BeadStepR) at a time, exactly like the beads on its own
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
	// Ring = 1 (no correction needed). Pair = 64 (runs at 1/64 the ring's clock for the
	// same slider setting). Never 0 or negative — EffectiveClockSpeed guards against a
	// division by an invalid value from an unrecognised or malformed scene.
	//
	// The pair's value is a WATCHED-IT number, not a derived one, and it has moved three
	// times against the live editor: 4 read too fast, 16 read too fast, 64 is where it
	// sits. Nothing computes it from the scene's size, so treat it as a setting to
	// re-watch rather than a constant with a proof behind it. Tests assert the CONSTRAINT
	// (ring unscaled, pair slower) and never this literal, precisely so re-watching it
	// costs one line here.
	ClockDivisor float64
	// DistanceGroups says whether this scene has the three named node-pair distance groups
	// (distance_groups.go's distanceGroups table) — the content of the "distance home
	// button" panel. The table names node ids ("1", "2", "3"…), and NODE IDS ARE PER-SCENE:
	// they are directory names under each tree's own nodes/, so the same id exists in every
	// scene and means a different node in each.
	//
	// Which is how the panel turned up in the PAIR tab. The groups are the RING's, and its
	// "input" group holds the pair (1, 2). The pair scene's two nodes are also called 1 and
	// 2, so that group resolved there — against two nodes it was never about — and the panel
	// showed a live length for it. The panel is data-driven and hides itself when every
	// group reads 0, so this flag is the whole fix: a scene without the groups streams three
	// zeroes and the panel does not render. No TS change, no scene name in the webview.
	//
	// An unknown tree (a fixture, a one-off path) gets false: a scene that is not the ring
	// has no claim on the ring's groups.
	DistanceGroups bool
	// Editable says whether this scene takes STRUCTURAL edits — the node palette's drop and
	// the delete key (scene_structure.go). A per-scene property for the same reason every
	// other field here is one: it is a fact about the scene, and Go owns it, so the editor
	// asks rather than branching on a scene NAME it would have to be told.
	//
	// Both scenes take edits. It was pair-only at first, on the reasoning that the ring is
	// the long-lived diagram whose layout has been tuned by hand across many sessions and
	// the pair is the small one built to be experimented with — but that argues for being
	// careful in the ring, not for the editor refusing to build one. What each scene accepts
	// is Kinds, below; whether it accepts anything at all is this.
	Editable bool
	// Kinds names the node kinds this scene ACCEPTS — what its palette offers and what a
	// create is allowed to make. Empty means every registered kind, which is what a scene
	// that has never thought about it gets.
	//
	// It is a per-scene list rather than a per-kind flag because "is this a pair kind" is not
	// a property of the kind: PairNode is the pair's node AND the ring's, and a kind that suits
	// two scenes should not have to name them. The scene knows what it is made of.
	Kinds []string
}

// SceneTabs is the tab strip, in display order. Index 0 is the DEFAULT: its Dir must be
// the anchor's own basename, since that is the path the extension host launches with and
// sizes its stream fds from (see AnchorIsTabbed, in scene_selection.go).
var SceneTabs = []SceneTab{
	// The ring takes the PIPELINE kinds — the nine it is already built from, plus HoldFlip
	// and Pacer, which speak the same vocabulary (a value in, a value on) and were simply
	// not used in this particular diagram. Not PairNode or NormalSum: those hold a tilt
	// vector and exchange directions, which is the pair's model, not a chain of gates.
	{Name: "ring", Dir: "topology", QuantizedDrag: false, CoplanarEdges: false, UpAxis: false, ClockDivisor: 1, DistanceGroups: true, Editable: true,
		Kinds: []string{
			"Input", "Time", "TimeStart", "TimeEnd",
			"Pulse", "PulseLeft", "PulseRight",
			"SelectLeft", "SelectRight",
			"HoldFlip", "Pacer",
		}},
	// The pair takes PAIR KINDS ONLY: PairNode, which is what both of its nodes are, and
	// NormalSum, which exists to sum two of their normals. Dropping a ring kind here would
	// build something the pair's own model has no place for — its exchange is two nodes and
	// the vectors between them, not a pipeline of gates.
	{Name: "pair", Dir: "topology-pair", QuantizedDrag: false, CoplanarEdges: true, UpAxis: true, ClockDivisor: 64, DistanceGroups: false, Editable: true, Kinds: []string{"PairNode", "NormalSum"}},
}
