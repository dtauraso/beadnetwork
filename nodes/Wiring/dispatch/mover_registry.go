// mover_registry.go — the nodeMover/edgeMover directory owner split out of MoveDispatch
// (god-object decomposition), as a pure move (no logic changes): moverRegistry owns
// nodeMovers/edgeMovers/edgeOut and the bind/start/sendMove/enqueueFor/
// centerOfNode logic. In-package callers address md.mr.X directly (bind/edgeOutFor/
// centerOfNode/enqueueFor/finalizeActors have no MoveDispatch-level delegator); only
// Start stays on MoveDispatch, since it also sets md.ctx. sendMove threads through
// md.ctx (owned elsewhere, NOT part of this extraction) as a parameter. The test-only
// message tap is per-mover (nodeactor.NodeGeometry's own tap field) — enqueueFor no
// longer takes or threads a shared tap, and (since §20's package move) no longer touches
// that field directly either: it forwards to NodeGeometry.EnqueueSend.
//
// The per-node actor (nodeGeometry/nodeMover) moved to package nodeactor in
// docs/planning/movedispatch-decomposition.md §20, the same package-boundary move §17 made
// for edgeMover — this file is the biggest concentration of the call sites that move
// updated to the exported nodeactor API.

package dispatch

import (
	"context"
	"fmt"
	"sync"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgemover"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// moverInboxDepth is the declared capacity of every per-mover movemsg.Msg inbox: an
// edgeMover's extIn/srcIn/dstIn (nodes/Wiring/edgemover), a nodeactor.NodeGeometry's extIn
// (nodeactor's own copy, nodeactor/consts.go), and each directed neighborIn/tiltEditIn/
// latticeIn built in this package (move_dispatch_construct.go, build_args_tilt_vector.go,
// build_args_lattice.go). Previously the same bare 8 repeated at six construction sites —
// the largest group of magic numbers in the network.
//
// WHY THIS NUMBER, honestly: it is a chosen depth, NOT derived. What IS derivable
// from the topology is the COUNT of these channels (one triple per edge, one extIn
// per node, two directed neighborIn per edge — 10/9/20 for the shipped graph), and
// the loader already fixes that by construction. The DEPTH is a queue for a burst of
// move messages during a drag, so it is bounded by gesture rate, not by anything in
// the spec files (docs/planning/visual-editor/session-log.md classifies it DYNAMIC for exactly this reason).
// 8 is "a few frames of drag messages"; it has held in practice and no measurement
// yet contradicts it.
const moverInboxDepth = 8

// moverRegistry is the pure registry that owns every mover and wires their dedicated
// channels together — there is no shared dispatch map; nodeMovers/edgeMovers themselves
// are the directories a mover's resolveDest closure and the external-entry helpers below
// look up. It also retains the per-edge source Outs so in-package callers can read an
// edge's loaded geometry (edgeOutFor) without going through a central coordinator.
type moverRegistry struct {
	// nodeGeoms is the UNIVERSAL per-node directory — every node's own
	// *nodeactor.NodeGeometry, ring and pair alike. This is what routing (resolveDest,
	// sendMove, centerOfNode, NodeKind, drag/commit) looks up: a message addressed to a
	// node must arrive and be handled regardless of which goroutine drives that node's
	// geometry.
	nodeGeoms map[string]*nodeactor.NodeGeometry
	// nodeMovers is the RING-ONLY actor directory: one entry per node whose OWN kind did
	// NOT claim BuildArgs.ClaimSelfDrive, populated by finalizeActors AFTER buildNodes has
	// run (so every ClaimSelfDrive call has already happened). Used ONLY by start() to
	// launch a dedicated goroutine per ring node — a PAIR node (PairNode) has NO entry
	// here at all, by construction, not by a flag that says "launch nothing for me".
	nodeMovers map[string]*nodeactor.NodeMover
	// selfDriveClaimed holds, for each node id whose OWN kind claimed
	// BuildArgs.ClaimSelfDrive at build time (PairNode — the pair scene), true. Written
	// ONCE per entry by ClaimSelfDrive (build_args_selfdrive.go), on the single-threaded build path,
	// before finalizeActors runs and before any goroutine exists. finalizeActors reads it
	// to decide which nodes get a nodeMover actor at all — an id present here gets NONE.
	selfDriveClaimed map[string]bool
	edgeMovers       map[string]*edgemover.EdgeMover
	// edgeOut: edgeId → source *Out, for read-only access by tests/verifiers.
	edgeOut map[string]*wire.Out
	// centerMirror is the DISPATCH goroutine's OWN mirror of every node's last-known
	// world center, kept current by messages from each node's own goroutine.
	// Seeded once at construction (newMoveDispatch, single-threaded
	// setup, from each node's load-time geom) so the first framing read has every
	// center before any push arrives, then kept current by drainCenterMirror pulling
	// each nodeGeometry's own centerOut channel (via PollCenter). Written and read ONLY
	// from the dispatch/ gesture goroutine (centerOfNode is, after the quantize call sites
	// moved to each node's own partnerCenters map, called only from that goroutine) — no
	// lock.
	centerMirror map[string]vec3
}

// bind wires the per-edge source Outs (keyed "source.sourceHandle" in outSink) and dest
// wires (slotReg, keyed "target.targetHandle") into each edgeMover. Call once after node
// construction.
func (mr *moverRegistry) bind(outSink map[string]*wire.Out, slotReg inputcodec.SlotRegistry) {
	for edgeID, em := range mr.edgeMovers {
		var o *wire.Out
		if oo, ok := outSink[em.SrcID()+"."+em.SrcHandle()]; ok {
			o = oo
			em.SetOut(oo)
			mr.edgeOut[edgeID] = oo
		}
		if pw, ok := slotReg[em.DstID()+"."+em.DstHandle()]; ok {
			em.SetDest(pw)
			// The SOURCE node also takes this wire, paired with the outTargets entry for
			// the same edge: the source node's own goroutine drives it (NodeMover.Run)
			// and reads its in-flight fractions to light its own chain
			// (docs/bead-model/beads-are-the-edge.md step 3). The wire is no longer driven by a
			// goroutine of its own — that is what "the wire goroutine is removed" means
			// concretely, and it is why the node can read the fraction without touching
			// another goroutine's state.
			// o may be nil if this edge's source handle wasn't found in outSink;
			// chainBeads then just skips publishing for this edge (it still lays the
			// chain out — the step count is computed locally either way, see
			// edgeStepCount). em.SendSteps is the second delivery chainBeads makes
			// alongside PublishSteps, so the edgeMover's own goroutine (which cannot
			// read the Out directly — see nodeOuts.outStepsIn's own doc comment) can
			// revise an in-flight bead's remaining travel against the same freshly
			// computed count.
			if srcNM, ok := mr.nodeGeoms[em.SrcID()]; ok {
				srcNM.AddOutWire(pw, em.DstID(), o, em.SendSteps)
			}
		}
	}
}

// start launches every mover's goroutine — ONE goroutine per node and ONE per edge, no
// dedicated sender/watcher goroutines (an earlier shared-outbox-plus-sender-goroutine
// design was removed: each mover's own run loop drains its own inbox AND retries its own
// pending sends, non-blockingly, every cycle).
//
// Returns a *sync.WaitGroup covering every launched goroutine, so a caller that wants a
// complete shutdown (main.go: "wait for everything, then close" — see
// the wait-for-everything-then-close change) can wg.Wait() on it after cancelling
// ctx. Both nm.Run and em.Run select on ctx.Done() at the top of their loop (their only
// blocking call is SleepCycle, which also selects on ctx), so cancel-to-return is one
// clock tick, worst case. Callers that don't care about shutdown completeness (most
// existing tests) can ignore the return value — start(ctx) alone still compiles and
// still launches every goroutine exactly as before.
func (mr *moverRegistry) start(ctx context.Context) *sync.WaitGroup {
	wg := new(sync.WaitGroup)
	// mr.nodeMovers holds ONLY ring nodes by construction (finalizeActors never builds
	// one for a node that claimed BuildArgs.ClaimSelfDrive) — there is nothing to skip
	// here, unlike the old selfDriven-flag check this replaced.
	for _, nm := range mr.nodeMovers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			nm.Run(ctx)
		}()
	}
	for _, em := range mr.edgeMovers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			em.Run(ctx)
		}()
	}
	return wg
}

// finalizeActors builds the RING actor directory (mr.nodeMovers) from mr.nodeGeoms, AFTER
// buildNodes has run every kind's own build func — which is when a pair kind calls
// BuildArgs.ClaimSelfDrive (build_args_selfdrive.go) and so is the earliest point "which
// nodes self-drive" is fully known. Every node id NOT in claimed gets wrapped in a
// nodeactor.NodeMover and a fresh speed channel (per-goroutine-clock.md "Delivery"),
// appended to speedSinks; every id IN claimed gets no NodeMover at all — nothing to skip
// launching later, by construction, not by a flag. clockSrc is copied into that node's own
// geometry.clk lazily by NodeMover.Run at its own goroutine start (mirrors every other
// per-goroutine clock use).
func (mr *moverRegistry) finalizeActors(speedSinks *[]chan float64) {
	mr.nodeMovers = map[string]*nodeactor.NodeMover{}
	for id, ng := range mr.nodeGeoms {
		if mr.selfDriveClaimed[id] {
			continue
		}
		nm := nodeactor.NewNodeMover(ng)
		if speedSinks != nil {
			nodeSpeedCh := make(chan float64, 1)
			nm.SetSpeedCh(nodeSpeedCh)
			*speedSinks = append(*speedSinks, nodeSpeedCh)
		}
		mr.nodeMovers[id] = nm
	}
}

// drainCenterMirror drains every node's own centerOut channel non-blockingly (via
// PollCenter), updating mr.centerMirror with whatever's newest for each node. Called
// before every dispatch-side framing read (centerOfNode) so those reads always see the
// latest pushed center. This is EVENTUALLY CONSISTENT (a read may be one push behind a
// node that just moved on its own goroutine) — acceptable for camera/framing reads, which
// is the only remaining caller class (see moverRegistry.centerMirror's doc comment).
// Must only be called from the dispatch/gesture goroutine — it is the sole
// reader of every node's centerOut channel.
func (mr *moverRegistry) drainCenterMirror() {
	if mr.centerMirror == nil {
		mr.centerMirror = map[string]vec3{}
	}
	for id, nm := range mr.nodeGeoms {
		if c, ok := nm.PollCenter(); ok {
			mr.centerMirror[id] = c
		}
	}
}

// centerOfNode returns the current world center for a node id by draining the center
// mirror (drainCenterMirror) and reading mr.centerMirror. Must only be called from the
// dispatch/gesture goroutine.
func (mr *moverRegistry) centerOfNode(id string) (vec3, bool) {
	mr.drainCenterMirror()
	c, ok := mr.centerMirror[id]
	return c, ok
}

// sendMove routes one movemsg.Msg to a node's dedicated external-entry channel (extIn), if
// the id is a known node. This is the EXTERNAL-caller path — RootMove (drag) and
// gesture.go's dragStart send — not a mover-to-mover send (those go through a node's
// own pending/flushPending onto its OWN dedicated channel, never through this
// function), so it has no owning geometry to fire a tap through — this bare path never
// fires the test-only tap (see nodeactor's own EnqueueSend doc comment; only EnqueueSend,
// a node's own send, does). Looks up mr.nodeGeoms (every node, ring and pair alike — a
// drag/select/hover addressed to a pair node must still arrive and be handled, on that
// node's own goroutine) — a read-only directory once construction finishes, safe to
// read from any goroutine. ctx is threaded through from MoveDispatch (not part of
// moverRegistry).
func (mr *moverRegistry) sendMove(ctx context.Context, id string, msg movemsg.Msg) {
	nm, ok := mr.nodeGeoms[id]
	if !ok {
		return
	}
	nm.SendExternal(ctx, msg)
}

// enqueueFor returns nm's own non-blocking send function — nm.EnqueueSend itself
// (nodeactor/node_geometry_accessors.go), the exact method value package Wiring's
// move_dispatch_construct.go binds as nm.WireMessaging(..., md.mr.enqueueFor(nm), ...).
// Kept as a named wrapper (rather than inlining nm.EnqueueSend at the one call site)
// purely so this file keeps documenting the whole enqueue story in one place, matching
// where the old inline closure body used to live before §20 folded that body into the
// actor itself (docs/planning/movedispatch-decomposition.md).
func (mr *moverRegistry) enqueueFor(nm *nodeactor.NodeGeometry) func(id string, msg movemsg.Msg) {
	return nm.EnqueueSend
}

// nodeKind returns the kind string for the given node id, or "" if unknown. Kind lives
// on nm.geom's embedded nodegeom.NodeIdentity — see MoveDispatch.NodeKind's former doc
// comment (move_dispatch_api.go, before this pure single-owner forward moved here) for
// why reading it off mr.nodeGeoms is safe under concurrent access.
func (mr *moverRegistry) nodeKind(nodeID string) string {
	if nm, ok := mr.nodeGeoms[nodeID]; ok {
		return nm.Kind()
	}
	return ""
}

// nodeBodyRadius is the node's body sphere radius used to size the home fit (gestHome,
// gesture_handlers.go). It reuses the SAME nodeRadius the pre-branch HomeButton framed
// with: nodegeom.NodeRadius(kind) = min(width,height)/CurveParamNodeRadiusDivisor, with
// the (110,60) default for an unknown kind.
func (mr *moverRegistry) nodeBodyRadius(id string) float64 {
	return nodegeom.NodeRadius(mr.nodeKind(id))
}

// hasNodeMover reports whether node id has a real, separate nodeactor.NodeMover actor (a
// ring node) as opposed to no NodeMover at all (a self-driven pair node, or an unknown id).
func (mr *moverRegistry) hasNodeMover(id string) bool {
	_, ok := mr.nodeMovers[id]
	return ok
}

// nodeSelfDriven reports whether node id is driven by its OWN pair-node goroutine
// (task/pair-node-owns-itself, ClaimSelfDrive) rather than a separate nodeactor.NodeMover
// goroutine — equivalently, whether id has NO entry in mr.nodeMovers at all
// (finalizeActors never builds one for a claimed id).
func (mr *moverRegistry) nodeSelfDriven(id string) bool {
	if _, hasGeom := mr.nodeGeoms[id]; !hasGeom {
		return false
	}
	return !mr.hasNodeMover(id)
}

// nodeQuantOffset returns node id's own current quantized polar offset triple
// (iTheta, iPhi, iR).
func (mr *moverRegistry) nodeQuantOffset(id string) (iTheta, iPhi, iR int, ok bool) {
	nm, exists := mr.nodeGeoms[id]
	if !exists {
		return 0, 0, 0, false
	}
	t, p, r := nm.QuantOffset()
	return t, p, r, true
}

// linkRefusal answers whether an edge from src to a NEW node of kind can exist, and says
// why not when it cannot. mr's only part is resolving src's own kind off nodeGeoms; the
// two structural reasons themselves are decided by the pure linkRefusalFor below.
func (mr *moverRegistry) linkRefusal(src, kind string) (srcPort, targetPort, why string, ok bool) {
	srcGeom, found := mr.nodeGeoms[src]
	srcKind := ""
	if found {
		srcKind = srcGeom.Kind()
	}
	return linkRefusalFor(src, srcKind, found, kind)
}

// linkRefusalFor is the pure decision linkRefusal makes once src's own kind (and whether
// it was found at all) has been resolved: kind must take an input, and src must have
// both geometry and an output to connect from. Split out of linkRefusal because it never
// touched moverRegistry itself, only the two node kinds it was handed.
func linkRefusalFor(src, srcKind string, srcFound bool, kind string) (srcPort, targetPort, why string, ok bool) {
	targetPort, hasIn := firstPortOfDir(kind, portwiring.PortIn)
	if !hasIn {
		return "", "", fmt.Sprintf("%s takes no input, so nothing can connect to it", kind), false
	}
	if !srcFound {
		return "", "", fmt.Sprintf("no geometry for %s", src), false
	}
	srcPort, hasOut := firstPortOfDir(srcKind, portwiring.PortOut)
	if !hasOut {
		return "", "", fmt.Sprintf("%s has no output to connect from", srcKind), false
	}
	return srcPort, targetPort, "", true
}

// nearestNodeTo picks the live node whose centre is closest to p, from this process's
// own geometry. The distance comparison itself is pure and lives in
// nodegeom.NearestTo; this method's only job is building the id->center map from mr's
// own nodeGeoms directory.
func (mr *moverRegistry) nearestNodeTo(p vec3) (string, bool) {
	centers := make(map[string]vec3, len(mr.nodeGeoms))
	for id, ng := range mr.nodeGeoms {
		centers[id] = ng.WorldCenter()
	}
	return nodegeom.NearestTo(centers, p)
}
