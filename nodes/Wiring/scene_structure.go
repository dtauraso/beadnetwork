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
	"math"
	"os"
	"strconv"

	B "github.com/dtauraso/wirefold/Buffer"
	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
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
		md.refuseStructuralEdit("this scene does not take structural edits")
		return
	}
	kind, ok := kindForID(kindID)
	if !ok {
		md.refuseStructuralEdit(fmt.Sprintf("unknown kind id %d", kindID))
		return
	}
	// A kind this SCENE does not take (SceneTab.Kinds). The palette does not offer it, so
	// this should be unreachable from the editor — which is exactly why it is checked: the
	// tree is written on this side, and "the UI does not offer it" is not "it cannot happen".
	if md.ui.sceneKinds&(1<<uint(kindID)) == 0 {
		md.refuseStructuralEdit(fmt.Sprintf("this scene does not take %s nodes", kind))
		return
	}
	// WHERE THE DROP LANDED. TS sent NDC; the camera that turns it into a place is Go's, so
	// the unprojection happens here — the same ray every node drag already unprojects
	// (dragPlaneHit), onto the camera-facing plane through the SCENE CENTRE. That plane,
	// rather than the plane through some node, is what makes a drop into empty space land
	// somewhere sensible: the scene sphere is the frame every node position is measured
	// from anyway.
	drop, okDrop := md.dropPointFromNDC(ndcX, ndcY)
	if !okDrop {
		md.refuseStructuralEdit("could not resolve where the drop landed")
		return
	}
	// The nearest node is also the SOURCE of the new edge: an edge is stored under its
	// source and carries no `source` key, so choosing the source is choosing the directory
	// the edge file lands in.
	src, okNear := md.nearestNodeTo(drop)
	target := newNodeID(md.Scenes.TreeRoot)
	var srcPort, targetPort string
	if okNear {
		var why string
		var canLink bool
		if srcPort, targetPort, why, canLink = md.linkRefusal(src, kind); !canLink {
			md.refuseStructuralEdit(why)
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
		md.refuseStructuralEdit(fmt.Sprintf("could not write node %s: %v", target, err))
		return
	}
	edges := countEdgeFiles(md.Scenes.TreeRoot)
	if okNear {
		if err := WriteEdgeFile(md.Scenes.TreeRoot, src, srcPort, target, targetPort); err != nil {
			md.refuseStructuralEdit(fmt.Sprintf("could not write edge %s->%s: %v", src, target, err))
			return
		}
		edges++
	}
	// An empty scene has no nearest node, so the new node stands alone. That is not an
	// error — there is nothing to refuse, only nothing to connect to.
	if err := WriteCounts(md.Scenes.TreeRoot, largestNodeID(md.Scenes.TreeRoot), edges); err != nil {
		md.refuseStructuralEdit(fmt.Sprintf("could not update counts.json: %v", err))
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
		md.refuseStructuralEdit("this scene does not take structural edits")
		return
	}
	id, ok := md.RT.LookupNodeRow(row)
	if !ok {
		md.refuseStructuralEdit(fmt.Sprintf("no node on row %d", row))
		return
	}
	root := md.Scenes.TreeRoot
	if err := RemoveNodeDir(root, id); err != nil {
		md.refuseStructuralEdit(fmt.Sprintf("could not remove node %s: %v", id, err))
		return
	}
	if err := RemoveEdgesTo(root, id, NodeIDStringsInTree(root)); err != nil {
		md.refuseStructuralEdit(fmt.Sprintf("could not remove edges into %s: %v", id, err))
		return
	}
	// The ROW SPACE KEEPS ITS HOLE. counts.json's "nodes" is the largest id, not a live
	// count, so deleting a middle node leaves its row empty rather than shifting the ids
	// above it down — that shift is the silent rename ROW ID = NODE ID - 1 exists to
	// prevent (node 6's geometry arriving on node 5's row the moment 5 is deleted).
	if err := WriteCounts(root, largestNodeID(root), countEdgeFiles(root)); err != nil {
		md.refuseStructuralEdit(fmt.Sprintf("could not update counts.json: %v", err))
		return
	}
	md.Scenes.Quit()
}

// linkRefusal answers whether an edge from src to a NEW node of kind can exist, and says why
// not when it cannot. Both reasons are structural facts Go already holds:
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
func (md *MoveDispatch) linkRefusal(src, kind string) (srcPort, targetPort, why string, ok bool) {
	targetPort, hasIn := firstPortOfDir(kind, PortIn)
	if !hasIn {
		return "", "", fmt.Sprintf("%s takes no input, so nothing can connect to it", kind), false
	}
	srcGeom, found := md.mr.nodeGeoms[src]
	if !found {
		return "", "", fmt.Sprintf("no geometry for %s", src), false
	}
	srcPort, hasOut := firstPortOfDir(srcGeom.geom.Kind, PortOut)
	if !hasOut {
		return "", "", fmt.Sprintf("%s has no output to connect from", srcGeom.geom.Kind), false
	}
	return srcPort, targetPort, "", true
}

// firstPortOfDir returns a registered kind's FIRST port in dir, in the order the kind
// declared them at RegisterBuilder. First, not "In": a kind names its own ports, and the
// declaration order is the only ranking there is — NormalSum's NormalA before NormalB says
// which one an edge should take when nothing else has been said.
func firstPortOfDir(kind string, dir PortDir) (string, bool) {
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

// dropPointFromNDC unprojects a drop's screen position onto the camera-facing plane through
// the SCENE CENTRE — the same ray-through-NDC a node drag already unprojects
// (gesture_actions.go's dragPlaneHit), against a plane that exists whether or not anything
// was under the pointer. ok=false when the ray is parallel to the plane or the hit is
// non-finite, which is a refusal rather than a guess at where the node should go.
func (md *MoveDispatch) dropPointFromNDC(ndcX, ndcY float64) (vec3, bool) {
	vp := md.ui.vp.Viewpoint
	eye := geom.EyeOf(vp)
	basis := geom.BasisFromViewpoint(vp.Pos, vp.Up)
	// g.fov/g.rect are the last render params the viewport reported. A gesture reads them
	// off the event it is handling; a palette DROP has no event of its own — it arrives as
	// an addressed edit, not raw input — so it uses the ones every pointer move across the
	// canvas has been keeping current.
	dir := geom.RayDirThroughNDC(ndcX, ndcY, basis, md.ui.gest.fov, md.ui.gest.rect.aspect())
	forward := basis.Pole.Scale(-1) // camera looks along -pole
	denom := dir.Dot(forward)
	if denom == 0 {
		return vec3{}, false
	}
	t := md.ui.sceneSphere.Center.Sub(eye).Dot(forward) / denom
	hit := eye.Add(dir.Scale(t))
	if math.IsNaN(hit.X) || math.IsInf(hit.X, 0) {
		return vec3{}, false
	}
	return hit, true
}

// nearestNodeTo picks the live node whose centre is closest to p, from this process's own
// geometry. Squared distance — the ordering is the same and there is no reason to take a
// square root to compare.
func (md *MoveDispatch) nearestNodeTo(p vec3) (string, bool) {
	best, bestD2, found := "", 0.0, false
	for id, ng := range md.mr.nodeGeoms {
		c := nodeWorldPos(ng.geom)
		d := c.Sub(p)
		d2 := d.X*d.X + d.Y*d.Y + d.Z*d.Z
		if !found || d2 < bestD2 {
			best, bestD2, found = id, d2, true
		}
	}
	return best, found
}

// refuseStructuralEdit reports a refused create/delete. It goes to STDERR, which the
// extension host pipes to the sim's output channel and .probe/go-errors.jsonl — the same
// route SelectScene's failed write takes, and where an operator looks when a gesture did
// nothing (memory/feedback_runner_errors_probe_first.md). Nothing is written and the run
// does not end, so the editor is exactly as it was.
func (md *MoveDispatch) refuseStructuralEdit(why string) {
	fmt.Fprintf(os.Stderr, "structural edit refused: %s\n", why)
	// …and SAY SO ON SCREEN. The reason belongs in the log; that the edit was refused at all
	// is the part a person cannot otherwise see, since the scene looks exactly as it did.
	// Bumping the count and emitting a frame is the whole signal — the editor watches the
	// number and shows a message when it goes up.
	md.ui.editRefused++
	md.emitViewFrame(nil)
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
// largestNodeID/nodeIDs/countEdgeFiles live in loader_tree.go, which is where reading the
// tree's shape belongs.
func newNodeID(root string) string {
	return strconv.Itoa(largestNodeID(root) + 1)
}

// THE WRITES THEMSELVES LIVE WITH THEIR OWNERS, not here: a per-node file is written by
// node_mover.go, a per-edge file by edge_mover.go, and the tree-level counts.json by
// loader_tree.go (check-persist-write-ownership, check-scene-path-resolution). This file
// decides WHAT happens and in what order; those files know where their own bytes go. The
// same split the rest of the persistence layer already has — an operation with no owner yet
// (a node that does not exist) still writes through the owner's path helpers.
