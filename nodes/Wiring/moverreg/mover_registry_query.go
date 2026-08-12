// mover_registry_query.go — MoverRegistry's read-only lookups: center mirror draining,
// external move routing, and the small per-node facts (kind, radius, mover presence,
// quant offset, nearest node) other dispatch code asks for. Split out of
// mover_registry.go by concern — see that file's header for the split rationale.
package moverreg

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
)

// drainCenterMirror drains every node's own centerOut channel non-blockingly (via
// PollCenter), updating mr.centerMirror with whatever's newest for each node. Called
// before every dispatch-side framing read (CenterOfNode) so those reads always see the
// latest pushed center. This is EVENTUALLY CONSISTENT (a read may be one push behind a
// node that just moved on its own goroutine) — acceptable for camera/framing reads, which
// is the only remaining caller class (see MoverRegistry.centerMirror's doc comment).
// Must only be called from the dispatch/gesture goroutine — it is the sole
// reader of every node's centerOut channel.
func (mr *MoverRegistry) drainCenterMirror() {
	if mr.centerMirror == nil {
		mr.centerMirror = map[string]vec3{}
	}
	for id, nm := range mr.nodeGeoms {
		if c, ok := nm.PollCenter(); ok {
			mr.centerMirror[id] = c
		}
	}
}

// CenterOfNode returns the current world center for a node id by draining the center
// mirror (drainCenterMirror) and reading mr.centerMirror. Must only be called from the
// dispatch/gesture goroutine.
func (mr *MoverRegistry) CenterOfNode(id string) (vec3, bool) {
	mr.drainCenterMirror()
	c, ok := mr.centerMirror[id]
	return c, ok
}

// SendMove routes one movemsg.Msg to a node's dedicated external-entry channel (extIn), if
// the id is a known node. This is the EXTERNAL-caller path — RootMove (drag) and
// gesture.go's dragStart send — not a mover-to-mover send (those go through a node's
// own pending/flushPending onto its OWN dedicated channel, never through this
// function), so it has no owning geometry to fire a tap through — this bare path never
// fires the test-only tap (see nodeactor's own EnqueueSend doc comment; only EnqueueSend,
// a node's own send, does). Looks up mr.nodeGeoms (every node, ring and pair alike — a
// drag/select/hover addressed to a pair node must still arrive and be handled, on that
// node's own goroutine) — a read-only directory once construction finishes, safe to
// read from any goroutine. ctx is threaded through from package dispatch's MoveDispatch
// (not part of MoverRegistry).
func (mr *MoverRegistry) SendMove(ctx context.Context, id string, msg movemsg.Msg) {
	nm, ok := mr.nodeGeoms[id]
	if !ok {
		return
	}
	nm.SendExternal(ctx, msg)
}

// EnqueueFor returns nm's own non-blocking send function — nm.EnqueueSend itself
// (nodeactor/node_geometry_accessors.go), the exact method value package dispatch's
// move_dispatch_construct.go binds as nm.WireMessaging(..., md.mr.EnqueueFor(nm), ...).
// Kept as a named wrapper (rather than inlining nm.EnqueueSend at the one call site)
// purely so this file keeps documenting the whole enqueue story in one place, matching
// where the old inline closure body used to live before §20 folded that body into the
// actor itself (docs/planning/movedispatch-decomposition.md).
func (mr *MoverRegistry) EnqueueFor(nm *nodeactor.NodeGeometry) func(id string, msg movemsg.Msg) {
	return nm.EnqueueSend
}

// nodeKind returns the kind string for the given node id, or "" if unknown. Kind lives
// on nm.geom's embedded nodegeom.NodeIdentity. Unexported: package dispatch has no
// external caller of this by itself, only through NodeBodyRadius/LinkRefusal below.
func (mr *MoverRegistry) nodeKind(nodeID string) string {
	if nm, ok := mr.nodeGeoms[nodeID]; ok {
		return nm.Kind()
	}
	return ""
}

// NodeBodyRadius is the node's body sphere radius used to size the home fit (gestHome,
// package dispatch's gesture_handlers.go). It reuses the SAME nodeRadius the pre-branch
// HomeButton framed with: nodegeom.NodeRadius(kind) =
// min(width,height)/CurveParamNodeRadiusDivisor, with the (110,60) default for an unknown
// kind.
func (mr *MoverRegistry) NodeBodyRadius(id string) float64 {
	return nodegeom.NodeRadius(mr.nodeKind(id))
}

// HasNodeMover reports whether node id has a real, separate nodeactor.NodeMover actor (a
// ring node) as opposed to no NodeMover at all (a self-driven pair node, or an unknown id).
func (mr *MoverRegistry) HasNodeMover(id string) bool {
	_, ok := mr.nodeMovers[id]
	return ok
}

// NodeSelfDriven reports whether node id is driven by its OWN pair-node goroutine
// (task/pair-node-owns-itself, ClaimSelfDrive) rather than a separate nodeactor.NodeMover
// goroutine — equivalently, whether id has NO entry in mr.nodeMovers at all
// (FinalizeActors never builds one for a claimed id).
func (mr *MoverRegistry) NodeSelfDriven(id string) bool {
	if _, hasGeom := mr.nodeGeoms[id]; !hasGeom {
		return false
	}
	return !mr.HasNodeMover(id)
}

// NodeQuantOffset returns node id's own current quantized polar offset triple
// (iTheta, iPhi, iR).
func (mr *MoverRegistry) NodeQuantOffset(id string) (iTheta, iPhi, iR int, ok bool) {
	nm, exists := mr.nodeGeoms[id]
	if !exists {
		return 0, 0, 0, false
	}
	t, p, r := nm.QuantOffset()
	return t, p, r, true
}

// NearestNodeTo picks the live node whose centre is closest to p, from this process's
// own geometry. The distance comparison itself is pure and lives in
// nodegeom.NearestTo; this method's only job is building the id->center map from mr's
// own nodeGeoms directory.
func (mr *MoverRegistry) NearestNodeTo(p vec3) (string, bool) {
	centers := make(map[string]vec3, len(mr.nodeGeoms))
	for id, ng := range mr.nodeGeoms {
		centers[id] = ng.WorldCenter()
	}
	return nodegeom.NearestTo(centers, p)
}
