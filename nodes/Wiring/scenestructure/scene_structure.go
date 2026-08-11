// Package scenestructure holds CREATING and DELETING a node, from the palette drop and the
// delete key — lifted out of nodes/Wiring/dispatch (docs/planning/movedispatch-decomposition.md
// §34). Neither operation belongs to any single existing owner: each reads/writes
// sceneswitch.SceneSwitch (the tree root + quit func), viewstate.UIState (scene-editable/
// scene-kinds gating, the drop-point unprojection, the scene sphere, the refusal counter +
// VIEW frame emit), moverreg.MoverRegistry (nearest-node/link-refusal), and
// rowtables.RowTables (row→id lookup for delete) — the same "genuinely its own boundary"
// shape distancegroups and sceneswitch itself were given their own package for.
//
// PERSIST, THEN END THE RUN. Neither operation touches the live network: a node's content
// stream is a dedicated fd the extension HOST allocates at spawn from counts.json (Node's
// spawn() takes its stdio array up front, before Go exists to be asked —
// .claude/rules/persistence-ownership.md), so a node created in-process would have no stream
// to emit on and an edge to it no way to draw. Both operations therefore write the tree and
// call the same quit the scene-tab switch uses; the host's looping runner respawns against
// the same anchor and the changed tree loads.
//
// That also makes these the ONLY writers of counts.json in the codebase — the file exists
// precisely because the host must size the stdio array before Go runs, and its own rule is
// single-writer: the operation that creates or deletes a node is the operation that updates
// it, and nothing else touches it.
package scenestructure

import (
	"fmt"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/countspersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgefile"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
	"github.com/dtauraso/wirefold/nodes/Wiring/sceneswitch"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

// CreateNode adds a node of kindID at a dropped world point, connected to the NEAREST
// existing node, then ends the run.
//
// The nearest node is resolved from Go's own live geometry. TS sends a POINT and never
// measures proximity — the same division every other edit follows (it sends which arrow was
// clicked, not what the new angle should be).
//
// A drop that cannot be connected is REFUSED: nothing is written, the run does not end, and
// the refusal is emitted so the editor can say so. A drop that silently does nothing is
// indistinguishable from a broken build.
func CreateNode(scenes *sceneswitch.SceneSwitch, ui *viewstate.UIState, mr *moverreg.MoverRegistry, kindID uint8, ndcX, ndcY float64, tr *T.Trace) {
	if scenes == nil || scenes.TreeRoot == "" || scenes.Quit == nil {
		return
	}
	// A scene that does not take structural edits refuses every one of them, here rather
	// than in the editor: the palette is hidden in such a scene, but "the UI does not offer
	// it" is not the same as "it cannot happen", and this is the side that owns the tree.
	if !ui.SceneEditable {
		ui.RefuseStructuralEdit("this scene does not take structural edits")
		ui.EmitViewFrame(nil)
		return
	}
	kind, ok := loadspec.KindForID(kindID)
	if !ok {
		ui.RefuseStructuralEdit(fmt.Sprintf("unknown kind id %d", kindID))
		ui.EmitViewFrame(nil)
		return
	}
	// A kind this SCENE does not take (SceneTab.Kinds). The palette does not offer it, so
	// this should be unreachable from the editor — which is exactly why it is checked: the
	// tree is written on this side, and "the UI does not offer it" is not "it cannot happen".
	if ui.SceneKinds&(1<<uint(kindID)) == 0 {
		ui.RefuseStructuralEdit(fmt.Sprintf("this scene does not take %s nodes", kind))
		ui.EmitViewFrame(nil)
		return
	}
	// WHERE THE DROP LANDED. TS sent NDC; the camera that turns it into a place is Go's, so
	// the unprojection happens here — the same ray every node drag already unprojects
	// (dragPlaneHit), onto the camera-facing plane through the SCENE CENTRE. That plane,
	// rather than the plane through some node, is what makes a drop into empty space land
	// somewhere sensible: the scene sphere is the frame every node position is measured
	// from anyway.
	drop, okDrop := ui.DropPointFromNDC(ndcX, ndcY)
	if !okDrop {
		ui.RefuseStructuralEdit("could not resolve where the drop landed")
		ui.EmitViewFrame(nil)
		return
	}
	// The nearest node is also the SOURCE of the new edge: an edge is stored under its
	// source and carries no `source` key, so choosing the source is choosing the directory
	// the edge file lands in.
	src, okNear := mr.NearestNodeTo(drop)
	target := loadspec.NewNodeID(scenes.TreeRoot)
	var srcPort, targetPort string
	if okNear {
		var why string
		var canLink bool
		if srcPort, targetPort, why, canLink = mr.LinkRefusal(src, kind); !canLink {
			ui.RefuseStructuralEdit(why)
			ui.EmitViewFrame(nil)
			return
		}
	}
	// The drop point is a WORLD point; a node's position is a SCENE POLAR about the scene
	// sphere (the fixed frame every node's position is measured from — polar-model.md), so
	// it is converted once, here, at the boundary. That is the same cartesian↔polar boundary
	// rule everything else follows: trig only where the two frames meet.
	// The scene centre comes off any node's own identity (every node carries the same
	// SceneCenter, established once at load — sphere_layout.go), read through the source
	// node this create already resolved. An empty scene has no node to read it from and no
	// nearest node either, so the drop is measured from the origin.
	c := ui.SceneSphere.Center
	off := drop.Sub(c)
	d := geom.WorldDirToAngles(off)
	if err := nodeactor.WriteNewNodeFiles(scenes.TreeRoot, target, kind, off.Length(), d.Theta, d.Phi); err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not write node %s: %v", target, err))
		ui.EmitViewFrame(nil)
		return
	}
	edges := loadspec.CountEdgeFiles(scenes.TreeRoot)
	if okNear {
		if err := edgefile.WriteEdgeFile(scenes.TreeRoot, src, srcPort, target, targetPort); err != nil {
			ui.RefuseStructuralEdit(fmt.Sprintf("could not write edge %s->%s: %v", src, target, err))
			ui.EmitViewFrame(nil)
			return
		}
		edges++
	}
	// An empty scene has no nearest node, so the new node stands alone. That is not an
	// error — there is nothing to refuse, only nothing to connect to.
	if err := countspersist.WriteCounts(scenes.TreeRoot, loadspec.LargestNodeID(scenes.TreeRoot), edges); err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not update counts.json: %v", err))
		ui.EmitViewFrame(nil)
		return
	}
	scenes.Quit()
}

// DeleteNode removes the node on a buffer ROW and EVERY edge touching it, then ends the run.
//
// Out-edges go with the directory. In-edges do not: an edge is stored only under its source
// (.claude/rules/persistence-ownership.md, "In-edges are deliberately not local"), so
// removing them is a pass over every other node's edges/ — the same walk the loader already
// makes. That cost is why in-edges are not duplicated, and paying it here is cheaper than
// keeping a second copy that can drift.
func DeleteNode(scenes *sceneswitch.SceneSwitch, ui *viewstate.UIState, rt *rowtables.RowTables, row int, tr *T.Trace) {
	if scenes == nil || scenes.TreeRoot == "" || scenes.Quit == nil {
		return
	}
	// A scene that does not take structural edits refuses every one of them, here rather
	// than in the editor: the palette is hidden in such a scene, but "the UI does not offer
	// it" is not the same as "it cannot happen", and this is the side that owns the tree.
	if !ui.SceneEditable {
		ui.RefuseStructuralEdit("this scene does not take structural edits")
		ui.EmitViewFrame(nil)
		return
	}
	id, ok := rt.LookupNodeRow(row)
	if !ok {
		ui.RefuseStructuralEdit(fmt.Sprintf("no node on row %d", row))
		ui.EmitViewFrame(nil)
		return
	}
	root := scenes.TreeRoot
	if err := nodeactor.RemoveNodeDir(root, id); err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not remove node %s: %v", id, err))
		ui.EmitViewFrame(nil)
		return
	}
	if err := edgefile.RemoveEdgesTo(root, id, loadspec.NodeIDStringsInTree(root)); err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not remove edges into %s: %v", id, err))
		ui.EmitViewFrame(nil)
		return
	}
	// The ROW SPACE KEEPS ITS HOLE. counts.json's "nodes" is the largest id, not a live
	// count, so deleting a middle node leaves its row empty rather than shifting the ids
	// above it down — that shift is the silent rename ROW ID = NODE ID - 1 exists to
	// prevent (node 6's geometry arriving on node 5's row the moment 5 is deleted).
	if err := countspersist.WriteCounts(root, loadspec.LargestNodeID(root), loadspec.CountEdgeFiles(root)); err != nil {
		ui.RefuseStructuralEdit(fmt.Sprintf("could not update counts.json: %v", err))
		ui.EmitViewFrame(nil)
		return
	}
	scenes.Quit()
}

// linkRefusal (moverRegistry.LinkRefusal, mover_registry.go) answers whether an edge from
// src to a NEW node of kind can exist, and says why not when it cannot. Both reasons are
// structural facts Go already holds:
//
//   - the new kind declares NO input port at all — nothing can be sent to it, so an edge to
//     it would be a line with no meaning;
//   - the source has no free OUTPUT port — every one it declares is already bound by an
//     existing edge.
//
// A kind's ports are its explicit []PortSpec at RegisterBuilder, which is why this can be
// answered without building anything.
// It also returns the PORTS the edge would use, because deciding a link is possible and
// deciding which ports carry it are the same question asked once. Answering them separately
// is what let an edge be written to a port that does not exist: the check looked at the
// kind's real ports, and the writer then assumed "Out" and "In".

// dropPointFromNDC (viewstate.UIState.DropPointFromNDC) unprojects a drop's screen
// position onto the camera-facing plane through the SCENE CENTRE — the same ray-through-NDC
// a node drag already unprojects (gesture_actions.go's dragPlaneHit), against a plane that
// exists whether or not anything was under the pointer. ok=false when the ray is parallel to
// the plane or the hit is non-finite, which is a refusal rather than a guess at where the
// node should go.

// nearestNodeTo (moverRegistry.NearestNodeTo, mover_registry.go) picks the live node whose
// centre is closest to p, from this process's own geometry.

// refuseStructuralEdit reports a refused create/delete. It goes to STDERR, which the
// extension host pipes to the sim's output channel and .probe/go-errors.jsonl — the same
// route SelectScene's failed write takes, and where an operator looks when a gesture did
// nothing (memory/feedback_runner_errors_probe_first.md). Nothing is written and the run
// does not end, so the editor is exactly as it was.
//
// It mutates ui's refusal counter only; every call site (CreateNode/DeleteNode's own
// refusal returns, in this file) follows it with ui.EmitViewFrame(nil) itself — the VIEW
// frame is emitted by the caller, per docs/planning/movedispatch-decomposition.md's
// write-then-emit split. Bumping the count and emitting a frame is the whole signal — the
// editor watches the number and shows a message when it goes up.
//
// Lives at viewstate.UIState.RefuseStructuralEdit (docs/planning/gesture-actor.md's lift) —
// call sites in this file read ui.RefuseStructuralEdit(...).

// kindForID/newNodeID live at loadspec.KindForID/loadspec.NewNodeID (god-object
// decomposition) — both are pure functions of their arguments with no Wiring state.
// LargestNodeID/NodeIDsInTree/CountEdgeFiles live in nodes/Wiring/loadspec's
// loader_tree.go, which is where reading the tree's shape belongs.

// THE WRITES THEMSELVES LIVE WITH THEIR OWNERS, not here: a per-node file is written by
// node_mover.go, a per-edge file by edge_mover.go, and the tree-level counts.json by
// loader_tree.go (check-persist-write-ownership, check-scene-path-resolution). This file
// decides WHAT happens and in what order; those files know where their own bytes go. The
// same split the rest of the persistence layer already has — an operation with no owner yet
// (a node that does not exist) still writes through the owner's path helpers.
