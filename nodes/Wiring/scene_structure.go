// scene_structure.go — CREATING and DELETING a node, from the palette drop and the delete
// key.
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
package Wiring

import (
	"fmt"
	"os"
	"strconv"

	B "github.com/dtauraso/wirefold/Buffer"
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/countspersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgefile"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
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
func (md *MoveDispatch) CreateNode(kindID uint8, ndcX, ndcY float64, tr *T.Trace) {
	if md == nil || md.Scenes.TreeRoot == "" || md.Scenes.Quit == nil {
		return
	}
	// A scene that does not take structural edits refuses every one of them, here rather
	// than in the editor: the palette is hidden in such a scene, but "the UI does not offer
	// it" is not the same as "it cannot happen", and this is the side that owns the tree.
	if !md.ui.sceneEditable {
		md.ui.refuseStructuralEdit("this scene does not take structural edits")
		md.emitViewFrame(nil)
		return
	}
	kind, ok := kindForID(kindID)
	if !ok {
		md.ui.refuseStructuralEdit(fmt.Sprintf("unknown kind id %d", kindID))
		md.emitViewFrame(nil)
		return
	}
	// A kind this SCENE does not take (SceneTab.Kinds). The palette does not offer it, so
	// this should be unreachable from the editor — which is exactly why it is checked: the
	// tree is written on this side, and "the UI does not offer it" is not "it cannot happen".
	if md.ui.sceneKinds&(1<<uint(kindID)) == 0 {
		md.ui.refuseStructuralEdit(fmt.Sprintf("this scene does not take %s nodes", kind))
		md.emitViewFrame(nil)
		return
	}
	// WHERE THE DROP LANDED. TS sent NDC; the camera that turns it into a place is Go's, so
	// the unprojection happens here — the same ray every node drag already unprojects
	// (dragPlaneHit), onto the camera-facing plane through the SCENE CENTRE. That plane,
	// rather than the plane through some node, is what makes a drop into empty space land
	// somewhere sensible: the scene sphere is the frame every node position is measured
	// from anyway.
	drop, okDrop := md.ui.dropPointFromNDC(ndcX, ndcY)
	if !okDrop {
		md.ui.refuseStructuralEdit("could not resolve where the drop landed")
		md.emitViewFrame(nil)
		return
	}
	// The nearest node is also the SOURCE of the new edge: an edge is stored under its
	// source and carries no `source` key, so choosing the source is choosing the directory
	// the edge file lands in.
	src, okNear := md.mr.nearestNodeTo(drop)
	target := newNodeID(md.Scenes.TreeRoot)
	var srcPort, targetPort string
	if okNear {
		var why string
		var canLink bool
		if srcPort, targetPort, why, canLink = md.mr.linkRefusal(src, kind); !canLink {
			md.ui.refuseStructuralEdit(why)
			md.emitViewFrame(nil)
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
	c := md.ui.sceneSphere.Center
	off := drop.Sub(c)
	d := geom.WorldDirToAngles(off)
	if err := WriteNewNodeFiles(md.Scenes.TreeRoot, target, kind, off.Length(), d.Theta, d.Phi); err != nil {
		md.ui.refuseStructuralEdit(fmt.Sprintf("could not write node %s: %v", target, err))
		md.emitViewFrame(nil)
		return
	}
	edges := loadspec.CountEdgeFiles(md.Scenes.TreeRoot)
	if okNear {
		if err := edgefile.WriteEdgeFile(md.Scenes.TreeRoot, src, srcPort, target, targetPort); err != nil {
			md.ui.refuseStructuralEdit(fmt.Sprintf("could not write edge %s->%s: %v", src, target, err))
			md.emitViewFrame(nil)
			return
		}
		edges++
	}
	// An empty scene has no nearest node, so the new node stands alone. That is not an
	// error — there is nothing to refuse, only nothing to connect to.
	if err := countspersist.WriteCounts(md.Scenes.TreeRoot, loadspec.LargestNodeID(md.Scenes.TreeRoot), edges); err != nil {
		md.ui.refuseStructuralEdit(fmt.Sprintf("could not update counts.json: %v", err))
		md.emitViewFrame(nil)
		return
	}
	md.Scenes.Quit()
}

// DeleteNode removes the node on a buffer ROW and EVERY edge touching it, then ends the run.
//
// Out-edges go with the directory. In-edges do not: an edge is stored only under its source
// (.claude/rules/persistence-ownership.md, "In-edges are deliberately not local"), so
// removing them is a pass over every other node's edges/ — the same walk the loader already
// makes. That cost is why in-edges are not duplicated, and paying it here is cheaper than
// keeping a second copy that can drift.
func (md *MoveDispatch) DeleteNode(row int, tr *T.Trace) {
	if md == nil || md.Scenes.TreeRoot == "" || md.Scenes.Quit == nil {
		return
	}
	// A scene that does not take structural edits refuses every one of them, here rather
	// than in the editor: the palette is hidden in such a scene, but "the UI does not offer
	// it" is not the same as "it cannot happen", and this is the side that owns the tree.
	if !md.ui.sceneEditable {
		md.ui.refuseStructuralEdit("this scene does not take structural edits")
		md.emitViewFrame(nil)
		return
	}
	id, ok := md.RT.LookupNodeRow(row)
	if !ok {
		md.ui.refuseStructuralEdit(fmt.Sprintf("no node on row %d", row))
		md.emitViewFrame(nil)
		return
	}
	root := md.Scenes.TreeRoot
	if err := RemoveNodeDir(root, id); err != nil {
		md.ui.refuseStructuralEdit(fmt.Sprintf("could not remove node %s: %v", id, err))
		md.emitViewFrame(nil)
		return
	}
	if err := edgefile.RemoveEdgesTo(root, id, loadspec.NodeIDStringsInTree(root)); err != nil {
		md.ui.refuseStructuralEdit(fmt.Sprintf("could not remove edges into %s: %v", id, err))
		md.emitViewFrame(nil)
		return
	}
	// The ROW SPACE KEEPS ITS HOLE. counts.json's "nodes" is the largest id, not a live
	// count, so deleting a middle node leaves its row empty rather than shifting the ids
	// above it down — that shift is the silent rename ROW ID = NODE ID - 1 exists to
	// prevent (node 6's geometry arriving on node 5's row the moment 5 is deleted).
	if err := countspersist.WriteCounts(root, loadspec.LargestNodeID(root), loadspec.CountEdgeFiles(root)); err != nil {
		md.ui.refuseStructuralEdit(fmt.Sprintf("could not update counts.json: %v", err))
		md.emitViewFrame(nil)
		return
	}
	md.Scenes.Quit()
}

// linkRefusal (moverRegistry.linkRefusal, mover_registry.go) answers whether an edge from
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

// firstPortOfDir returns a registered kind's FIRST port in dir, in the order the kind
// declared them at RegisterBuilder. First, not "In": a kind names its own ports, and the
// declaration order is the only ranking there is — NormalSum's NormalA before NormalB says
// which one an edge should take when nothing else has been said.
func firstPortOfDir(kind string, dir portwiring.PortDir) (string, bool) {
	b, ok := Registry[kind]
	if !ok {
		return "", false
	}
	for _, p := range b.Ports {
		if p.Dir == dir {
			return p.Name, true
		}
	}
	return "", false
}

// dropPointFromNDC (uiState.dropPointFromNDC, ui_state.go) unprojects a drop's screen
// position onto the camera-facing plane through the SCENE CENTRE — the same ray-through-NDC
// a node drag already unprojects (gesture_actions.go's dragPlaneHit), against a plane that
// exists whether or not anything was under the pointer. ok=false when the ray is parallel to
// the plane or the hit is non-finite, which is a refusal rather than a guess at where the
// node should go.

// nearestNodeTo (moverRegistry.nearestNodeTo, mover_registry.go) picks the live node whose
// centre is closest to p, from this process's own geometry.

// refuseStructuralEdit reports a refused create/delete. It goes to STDERR, which the
// extension host pipes to the sim's output channel and .probe/go-errors.jsonl — the same
// route SelectScene's failed write takes, and where an operator looks when a gesture did
// nothing (memory/feedback_runner_errors_probe_first.md). Nothing is written and the run
// does not end, so the editor is exactly as it was.
//
// It mutates ui's refusal counter only; every call site (CreateNode/DeleteNode's own
// refusal returns, in this file) follows it with md.emitViewFrame(nil) itself — the VIEW
// frame is emitted by the caller, per docs/planning/movedispatch-decomposition.md's
// write-then-emit split. Bumping the count and emitting a frame is the whole signal — the
// editor watches the number and shows a message when it goes up.
func (ui *uiState) refuseStructuralEdit(why string) {
	fmt.Fprintf(os.Stderr, "structural edit refused: %s\n", why)
	// …and SAY SO ON SCREEN. The reason belongs in the log; that the edit was refused at all
	// is the part a person cannot otherwise see, since the scene looks exactly as it did.
	ui.editRefused++
}

// kindForID reverses Buffer's kind-id map: the wire carries the numeric kind identity the
// Node block's KindId column already uses, so no kind NAME crosses the bridge.
func kindForID(id uint8) (string, bool) {
	for _, k := range B.KnownKinds() {
		if B.NodeKindID(k) == id {
			return k, true
		}
	}
	return "", false
}

// newNodeID is one past the largest id — never a reused hole. Reusing a freed id would make
// a node's identity ambiguous across a session boundary: the same directory name would name
// a different node before and after, which is the whole reason ids are not renumbered.
// LargestNodeID/NodeIDsInTree/CountEdgeFiles live in nodes/Wiring/loadspec's
// loader_tree.go, which is where reading the tree's shape belongs.
func newNodeID(root string) string {
	return strconv.Itoa(loadspec.LargestNodeID(root) + 1)
}

// THE WRITES THEMSELVES LIVE WITH THEIR OWNERS, not here: a per-node file is written by
// node_mover.go, a per-edge file by edge_mover.go, and the tree-level counts.json by
// loader_tree.go (check-persist-write-ownership, check-scene-path-resolution). This file
// decides WHAT happens and in what order; those files know where their own bytes go. The
// same split the rest of the persistence layer already has — an operation with no owner yet
// (a node that does not exist) still writes through the owner's path helpers.
