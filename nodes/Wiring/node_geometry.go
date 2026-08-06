// node_geometry.go — nodeGeometry: what a node's geometry actually IS (position, reach,
// kind, neighbour centres, outgoing wires, its own retry queue, tilt/received mirror
// fields, persist root, trace, dedicated stream writer) and every method that acts on it
// (task/pair-node-owns-itself — separating STATE+BEHAVIOUR from the ACTOR that drives it).
//
// A nodeGeometry is NOT a goroutine and NOT an inbox-draining actor by itself — it owns its
// own dedicated inbound channels (extIn/neighborIn) because a node's messages are part of
// what a node IS, but nothing here ever blocks on them or loops. Two different things drive
// a nodeGeometry's per-cycle work today:
//
//   - a RING node's dedicated nodeMover actor (node_mover.go: its own goroutine, launched
//     by moverRegistry.start), which owns a *nodeGeometry and paces it on its own clock; or
//   - a PAIR node's own kind goroutine (Node1/Node2, via BuildArgs.ClaimSelfDrive), which
//     owns a *nodeGeometry DIRECTLY — there is no nodeMover for it at all, no second
//     goroutine, nothing to skip launching.
//
// Either way exactly ONE goroutine ever touches a given nodeGeometry's mutable state — the
// invariant node_mover.go's own doc comments state throughout, unchanged by this split.
package Wiring

import (
	"encoding/binary"
	"fmt"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"math"

	T "github.com/dtauraso/wirefold/Trace"
)

// nodeGeometry owns one node's geometry, its own inbound channel set (extIn + one per
// neighbor — there is no single shared inbox), and its own outbound retry queue. On a move
// for itself it updates its held position and re-emits its node-geometry.
type nodeGeometry struct {
	id   string
	geom nodeGeom
	// persistRoot is the tree root this node writes its OWN per-node files (position.json
	// — quant_offset_persist.go; port anchor files — scene_anchor_persist.go) into. Set
	// once, for every node, by MoveDispatch.EnableEditPersist after the startup seed.
	// Empty ("") means unarmed — bare test construction, or a MoveDispatch built without
	// EnableEditPersist — and every persist* method below is a no-op. This node's own
	// goroutine (whichever one drives it) reads it only from its own persist* methods, so
	// no synchronization is needed even though every node shares the same
	// EnableEditPersist call that sets it (a plain string write before any driving
	// goroutine starts).
	persistRoot string
	// extIn is this node's dedicated channel for EXTERNAL entries — the stdin/gesture
	// goroutine's drag/dragStart sends (md.sendMove). Nothing else ever writes here: no
	// other node shares it.
	extIn chan moveMsg
	// neighborIn holds one dedicated inbound channel PER ADJACENT NODE (keyed by that
	// neighbor's id) — the "two channels, A→B and B→A" topology generalized to this
	// node's whole neighbor set. Built once at construction (newMoveDispatch) from edge
	// adjacency and never mutated afterward, so it's safe for the driving goroutine to
	// snapshot into a fixed select-case list at its own start. A neighbor M's own
	// goroutine is the only writer of neighborIn[M]; nothing else ever sends on it.
	neighborIn map[string]chan moveMsg
	tr         *T.Trace
	// clockSrc is the Clock this node's driving goroutine Copies from EXACTLY ONCE, at its
	// own start, into clk below (per-goroutine-clock.md). Set once at construction. Not
	// read again after that copy.
	clockSrc wire.Clock
	// clk is this node's OWN clock copy — read by writeStreamFrame (the frame tick) and,
	// for a ring node, by its owning nodeMover's pacing loop (ApplySpeedNonBlocking/
	// SleepCycle). Only the one goroutine driving this geometry ever reads or writes it.
	// Defaults to a fresh, real, live-ticking RealClock (see newNodeGeometry) so a test
	// that never launches a driving goroutine (e.g. a bare literal calling flushPending
	// directly) never dereferences a nil Clock.
	clk wire.Clock
	// speedCh is not here because polling one every cycle is pacing — an ACTOR concern. It
	// lives on whatever drives this geometry: nodeMover for a ring node (node_mover.go),
	// PairNodeSelf for a pair node (pair_node_self.go). BOTH must poll it.
	//
	// This comment used to say a pair node needed no such channel, because "its own kind
	// goroutine paces itself on its own clock already". That was wrong, and it hid a real
	// defect: a pair node has TWO clocks — its kind loop's, which is scaled, and THIS one.
	// This clock is what chainBeads (chain_beads.go) reads to lay out the bead animation
	// and what writeStreamFrame stamps a frame with, so while it went unscaled the pair
	// scene's VISIBLE motion ignored both the speed slider and SceneTab.ClockDivisor,
	// even though bead delivery timing was scaled correctly and looked fine.

	// There is no geomMu. m.geom (port_geometry.go) splits into an embedded, write-once
	// nodeIdentity (Kind/Label/R/SceneCenter — set once at construction in loader.go,
	// grepped clean of any later write anywhere in this package) and MUTABLE state
	// (ScenePolar/HasPos/ReachR) written only by applyCenter. Every writer AND every
	// reader of the mutable part — applyCenter, emitGeometry's full-struct copy — runs
	// exclusively on this node's OWN driving goroutine (whichever one that is), so there
	// is never more than one goroutine touching that memory. The one cross-goroutine
	// reader, MoveDispatch.NodeKind (node_move.go), called from the gesture/stdin-reader
	// goroutine, reads ONLY nm.geom.Kind — a field on the embedded nodeIdentity, which no
	// writer here ever touches.
	//
	// CHECKED BY CODE: TestNodeKindConcurrentWithApplyCenterUnderRace
	// (node_mover_geom_race_test.go) drives NodeKind's reader loop and applyCenter's
	// writer loop concurrently under -race, as a standing regression check that the split
	// holds.
	// centerOut is this node's OWN dedicated one-slot delivery channel to the DISPATCH
	// goroutine's owned center mirror (moverRegistry.centerMirror). A size-1 buffered
	// channel written with LATEST-WINS semantics (applyCenter drains any stale unread
	// value before sending the fresh one, never blocking): only the newest pushed center
	// matters to a framing read, so an unread stale value is simply overwritten rather
	// than queued. Only this node's own driving goroutine (applyCenter) ever sends here;
	// only the dispatch goroutine (moverRegistry.drainCenterMirror) ever receives.
	centerOut chan vec3
	// sendMove routes a moveMsg to another id's OWN dedicated channel (resolveDest,
	// above) — no shared inbox, no shared mutable state. Bound to md.enqueueFor(this): it
	// appends to pending and immediately attempts a non-blocking flush (never blocks the
	// calling handler goroutine).
	sendMove func(id string, msg moveMsg)
	edgeIDs  []string
	// centerOf resolves another node's current world center, bound to md.centerOfNode.
	// Unused by any live handler now that the rule/gate/anchor cascade (which used it to
	// read rule-neighbor centers) is gone; kept wired for any future direct-neighbor
	// lookup need.
	centerOf func(id string) (vec3, bool)
	// commitLocal is the OWNER-GOROUTINE commit path, bound to md.commitNodeMoveLocal
	// (generalized to every node). It publishes this node's own snap SYNCHRONOUSLY via
	// applyCenter instead of enqueuing an async self-send, so it is safe to call from
	// THIS node's own handle() for a moveMsgKindDrag, with no cross-goroutine self-send
	// and no shared mutable state (each node's quantized offset lives on its own
	// geometry — see quantOffset). nil in tests that build a bare nodeGeometry directly.
	commitLocal func(id string, newPos vec3)
	// partnerCenters is THIS node's OWN copy of every direct neighbor's last-known world
	// center, read by quantized_move.go's neighbor-move math. Written ONLY by this
	// node's own driving goroutine: seeded once at construction (newMoveDispatch,
	// single-threaded setup) from each neighbor's load-time geom, then kept current by
	// the moveMsgKindNeighborCenter handler in handle() below, fed by every direct
	// neighbor's own applyCenter push. Never read or written by any other goroutine.
	partnerCenters map[string]vec3
	// quantOffset is THIS node's own quantized polar offset (iTheta,iPhi,iR + step
	// constants) about the scene center. Seeded at load (buildMoveDispatch) from the
	// computed/persisted offset, then mutated ONLY by this node's own commit path
	// (commitNodeMoveCommon, called from this node's own driving goroutine via
	// commitLocal) — single-writer, no map, no race.
	quantOffset quantizedOffset
	// pending is THIS node's own outbound retry queue: sendMove appends here and
	// attempts an immediate non-blocking send; an item that can't be delivered right now
	// (the target's inbox is momentarily full) stays here and is retried — before any
	// newer item to the SAME destination — on the next flushPending call, which the
	// driving goroutine makes every cycle. There is no dedicated sender goroutine: only
	// this node's own driving goroutine ever touches pending (every sendMove call
	// originates from handle, which only ever runs on that same goroutine).
	pending []pendingSend
	// tap is a TEST-ONLY observability seam: when non-nil, THIS node's own enqueueFor
	// closure invokes it with every (destID, msg) it routes, before appending to
	// pending. nil in production.
	tap func(destID string, msg moveMsg)
	// resolveDest looks up the ONE dedicated directed channel FROM this node TO the given
	// destination id — the destination's neighborIn[this node's id] if destID is another
	// node, or the destination edge's srcIn/dstIn depending on which endpoint this node
	// is. There is no shared inbox to look up. nil only in tests that build a bare
	// nodeGeometry directly, in which case flushPending is a no-op.
	resolveDest func(id string) (chan moveMsg, bool)

	// --- chain bead actors (bead_chain.go) ---
	beadTickFn func() <-chan struct{}
	beadChains map[string]*edgeBeadChain

	// --- dedicated per-node stream (memory/feedback_no_single_writer_bridge.md) ---
	streamOut claimedStream
	nodeRow   int32
	selfKind  string

	outTargets     []string
	outWires       []*wire.PacedWire
	outWireTargets []string
	outWireOuts    []*wire.Out
	outStepsIn     []chan int

	neighborKinds map[string]string
	mutualTargets map[string]bool
	coplanarEdges bool
	upAxis        bool
	nodeRowFor    func(id string) (int32, bool)

	// --- own selection/hover/abc-drag UI state (per-owner, no shared/republished map) ---
	selected, hovered, latchedSel uint8
	hoverPort                     string
	hoverIsInput                  bool
	kindID                        uint8

	buildFrame func(tick uint32, nodeRow int32, nodeID int32, cx, cy, cz, radius, sphereR float32, vrx, vry, vrz, frx, fry, frz float32, poleTheta, polePhi, ringAxisTheta, ringAxisPhi, topTiltVectorLen, topTiltVectorTheta, topTiltVectorPhi, bottomTiltVectorTheta, bottomTiltVectorPhi, coplanarNormalTheta, coplanarNormalPhi, receivedVectorLen, receivedVectorTheta, receivedVectorPhi float32, selected, kindID, hovered, latchedSel uint8, label string, chainBeadOX, chainBeadOY, chainBeadOZ []float32, chainBeadLit []uint8, chainBeadLitValue []int32, events []wire.RowEvent) []byte

	topTiltVectorThetaIdx, topTiltVectorPhiIdx   int32
	normalThetaIdx, normalPhiIdx                 int32
	bottomThetaIdx, bottomPhiIdx                 int32
	receivedVectorThetaIdx, receivedVectorPhiIdx int32
	receivedVectorSet                            bool
}

// newNodeGeometry constructs one node's geometry — no actor, no goroutine. Whoever drives
// it (a ring's nodeMover, or a pair kind's own goroutine via ClaimSelfDrive) copies
// clockSrc into clk once, at its own start.
func newNodeGeometry(id string, geom nodeGeom, tr *T.Trace, clockSrc wire.Clock) *nodeGeometry {
	ng := &nodeGeometry{
		id: id, geom: geom,
		extIn: make(chan moveMsg, moverInboxDepth), neighborIn: map[string]chan moveMsg{}, tr: tr,
		partnerCenters: map[string]vec3{},
		centerOut:      make(chan vec3, 1),
		clockSrc:       clockSrc, clk: wire.NewRealClock(),
	}
	// Self-seed centerOut with the initial geometry (even when !HasPos, in which case
	// nodeWorldPos falls back to the origin) so the dispatch goroutine's first drain
	// always finds a valid center.
	ng.centerOut <- nodeWorldPos(geom)
	// Production-only hook: arms the bead-actor path in chainBeads/reconcileBeadChain
	// (bead_chain.go). Bare `&nodeGeometry{...}` test literals never call
	// newNodeGeometry, so beadTickFn stays nil there and chainBeads' pure-function tests
	// never touch a live TickBroadcaster goroutine.
	ng.beadTickFn = wire.NewTickChan
	return ng
}

// handle applies one move to this node: update held position, re-emit node-geometry.
func (m *nodeGeometry) handle(msg moveMsg) {
	if msg.NodeID != m.id {
		return
	}
	if msg.Kind == moveMsgKindCenter {
		// This node is the SOLE writer of its own position (single-writer by
		// construction — this is the only path that mutates it). A Center payload is
		// the flat absolute-scene-polar drag write from fanCenters: apply it via
		// applyCenter, which also re-emits. A nil Center is fanCenters' PARTNER
		// re-emit (a neighbor whose OWN center is unchanged, only asked to re-emit so
		// any reader of its geometry sees a consistent frame) — no mutation, just
		// re-emit.
		if msg.Center != nil {
			m.applyCenter(*msg.Center, msg.ReachR)
			return
		}
		if m.tr != nil {
			m.emitGeometry()
		}
		return
	}
	if msg.Kind == moveMsgKindDrag {
		// Owner-goroutine drag entry (generalized to EVERY node so no node's quantized
		// offset is ever touched by a foreign goroutine): commit this node's OWN new
		// position via the local (synchronous-snap-publish) commit path. A drag is
		// always a FREE move now -- there is no equal-radii solve and no propagation
		// past this node's own commit.
		newPos := msg.Target
		if m.commitLocal != nil {
			m.commitLocal(m.id, newPos)
		}
		if m.tr != nil {
			m.tr.Breadcrumb("drag.commit", m.id, "", fmt.Sprintf("newPos=(%.4f,%.4f,%.4f)", newPos.X, newPos.Y, newPos.Z))
			// Structured buffer counterpart, riding this node's own dedicated
			// stream frame (emitGeometry's own next emit already fires from
			// commitLocal above, so this rides as a distinct events-only-shaped
			// write here rather than waiting on that one).
			m.writeStreamFrame([]wire.RowEvent{{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbDragCommit, Debug: 1,
				NodeRow: m.nodeRow, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
				X: newPos.X, Y: newPos.Y, Z: newPos.Z,
			}})
		}
		return
	}
	if msg.Kind == moveMsgKindDragStart {
		m.startBeadDrag()
		return
	}
	if msg.Kind == moveMsgKindDragEnd {
		// "done dragging" is not optional (PLAN.md): sent from gesture.go's
		// gestPointerUp on EVERY path a drag can end by (including one abandoned
		// without a clean pointer-move first — see moveMsgKindDragEnd's own doc
		// comment), so no chain bead this node woke is ever left on machine time.
		m.endBeadDrag()
		return
	}
	if msg.Kind == moveMsgKindSelect {
		if msg.Bool {
			m.selected = 1
		} else {
			m.selected = 0
		}
		return
	}
	if msg.Kind == moveMsgKindHover {
		if msg.Bool {
			m.hovered = 1
			m.hoverPort = msg.Port
			m.hoverIsInput = msg.IsInput
		} else {
			m.hovered = 0
			m.hoverPort = ""
			m.hoverIsInput = false
		}
		return
	}
	if msg.Kind == moveMsgKindLatched {
		if msg.Bool {
			m.latchedSel = 1
		} else {
			m.latchedSel = 0
		}
		return
	}
	if msg.Kind == moveMsgKindTiltVectorAngle {
		// Adjust THIS node's own vector-direction index by one TiltVectorAngleStep click —
		// index arithmetic only (memory/feedback_abc_times_constant_not_rederive.md), no
		// trig here. Persisted immediately to this node's OWN file (persistTiltVectorAngle,
		// quant_offset_persist.go) and re-emitted so the panel's read-only reflect and
		// the drawn arrow both pick up the change on the next frame.
		delta := int32(-1)
		if msg.Bool {
			delta = 1
		}
		if msg.Axis == "phi" {
			m.topTiltVectorPhiIdx += delta
		} else {
			m.topTiltVectorThetaIdx += delta
		}
		m.persistTiltVectorAngle()
		if m.tr != nil {
			m.emitGeometry()
		}
		// NOTE: this path only runs for a kind that has NOT claimed BuildArgs.TiltEditIn
		// (every kind except Node1/Node2 today — see moveMsgKindTiltVectorAngle's own doc
		// comment and applyUpdateTiltVector's fallback, stdin_reader.go). Node1/Node2's own
		// tilt-panel edits are routed to their OWN goroutine instead (TiltEditIn), which
		// applies the click, syncs this value back via PairNodeSelf.SetTiltIndex, AND places
		// "the kick" bead on its own Out directly — none of that happens here anymore.
		return
	}
	if msg.Kind == moveMsgKindTiltVectorReset {
		// Return THIS node's own vector direction to the start position — both indices to
		// 0, the documented default (tilt vector at world +y). No bead: this is a
		// stop-and-return, not a kick. Persisted immediately, same as an adjust.
		m.topTiltVectorThetaIdx = 0
		m.topTiltVectorPhiIdx = 0
		m.persistTiltVectorAngle()
		if m.tr != nil {
			m.emitGeometry()
		}
		// NOTE: same split as moveMsgKindTiltVectorAngle — this path only runs for a kind
		// that has NOT claimed BuildArgs.TiltEditIn. Node1/Node2 route a reset through
		// their own TiltEditIn/TiltEditMsg.Reset instead.
		return
	}
	// moveMsgKindTiltIndexSync/ReceivedVectorSync/BeadClear are GONE
	// (task/pair-node-owns-itself): a pair node (Node1/Node2) owns this geometry
	// directly (PairNodeSelf, pair_node_self.go), so what used to be a one-way
	// notification message to itself is now a plain method call on the same
	// goroutine — see PairNodeSelf.SetTiltIndex/SetReceivedVector/ClearOutBeads,
	// which apply exactly what this handle() branch used to apply.
	if msg.Kind == moveMsgKindNeighborCenter {
		// Delivery-mechanism push (see applyCenter/partnerCenters' doc comments): a
		// direct neighbor's OWN center just changed. Store it in THIS node's owned
		// partnerCenters map (write, own goroutine only) and re-emit THIS node's own
		// geometry so its aimed ports pick up the fresh partner center — same value,
		// same effect as the old cross-goroutine snap read, just message-delivered.
		// ONE HOP ONLY: this node's own center did NOT change, so it must never push
		// a NeighborCenter of its own onward from here (no cascade past this point).
		if m.partnerCenters == nil {
			m.partnerCenters = map[string]vec3{}
		}
		m.partnerCenters[msg.SenderID] = msg.FromCenter
		if m.tr != nil {
			// DIAGNOSTIC ONLY (task/log-node4-chain-aim): records that this node's own
			// goroutine received a neighbor-center push, and from whom, so a drag-time
			// trace can show whether/when it arrives relative to this node's own emits.
			value := fmt.Sprintf("sender=%s center=(%.4f,%.4f,%.4f)", msg.SenderID, msg.FromCenter.X, msg.FromCenter.Y, msg.FromCenter.Z)
			m.tr.Breadcrumb("neighbor-center-recv", m.id, msg.SenderID, value)
			senderRow := int32(-1)
			if m.nodeRowFor != nil {
				if r, ok := m.nodeRowFor(msg.SenderID); ok {
					senderRow = r
				}
			}
			m.writeStreamFrame([]wire.RowEvent{{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbNeighborCenterRecv, Debug: 1,
				NodeRow: m.nodeRow, PortRow: -1, TargetRow: senderRow, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
				Text: value,
			}})
			m.emitGeometry()
		}
		return
	}
	if m.tr != nil {
		m.emitGeometry()
	}
}

// PerpendicularThetaIdx is the topTiltVectorThetaIdx value at which the tilt vector is exactly
// perpendicular to world +y: CurveParamTiltVectorAngleStep is π/12 (15°), and π/2 (90°) is
// exactly 6 steps. Comparing to this INTEGER is what makes the straightening loop's stop
// condition exact — cos(π/2) in float64 is ~6.1e-17, so a literal float dot==0 test would
// never fire (memory/feedback_abc_times_constant_not_rederive.md: index arithmetic, trig
// only at the cartesian/polar boundary). Exported (capitalized) so Node1/Node2's own
// goroutine — which now runs the straightening rule itself, per-package — can compare
// against it without duplicating the constant; the rule itself no longer lives here (see
// nodes/Node1/node.go, nodes/Node2/node.go).
//
// dot(tilt, coplanarNormal) == 0 is decided as thetaIdx == PerpendicularThetaIdx, not by
// computing an actual float dot product. STATE THE ASSUMPTION THAT MAKES THE SHORTCUT
// VALID: the tilt vector's in-plane angle IS its θ index only because, for this scene, the
// ring plane holds world +y and θ is measured from +y, so the two coincide (see
// topTiltVectorThetaIdx's own doc comment and the CoplanarNormal/UpAxis derivations in
// writeStreamFrame above). A scene whose ring plane does NOT contain +y breaks that
// coincidence — θ would then measure something unrelated to the coplanar normal, and the
// rule would need to compare an actual dot(tilt, coplanarNormal) via the two integer
// indices' angles converted through anglesToWorldOffset, not thetaIdx alone.
const PerpendicularThetaIdx int32 = 6

// repositionForTiltIndex implements the PAIR tab's model (task/tilt-sets-pair-distance,
// narrowed by task/node1-turns-every-round): the user asked ONLY for the pair's
// centre-to-centre DISTANCE to follow this node's own tilt-vector theta INDEX — nothing
// about ORIENTATION. Whenever this node's own goroutine reports a new index
// (PairNodeSelf.SetTiltIndex above), THIS node's own geometry changes ONLY its own
// quantized r-index (m.quantOffset.iR); iTheta and iPhi — which pin this node's own ray
// from the scene centre — are left EXACTLY as they are. There is no cartesian "direction"
// computed here (the earlier version picked the partner-to-self direction and committed a
// world point, which silently re-placed the node's ray and so changed the graph's
// ORIENTATION — a placement decision nobody asked for). The node simply slides along the
// ray it is already on.
//
// The distance this slide targets is
//
//	D = (abs(thetaIdx) + nodeTorusSteps(selfKind) + nodeTorusSteps(partnerKind)) * wire.BeadStepR
//
// so that edgeStepCount(D, srcKind, dstKind) (chain_beads.go) — never modified here — comes
// back out to exactly abs(thetaIdx).
//
// Because the pair is not exactly radial (the two members' loaded iPhi commonly differ by
// one step), this node's new iR is not simply "partner's iR minus steps" — it is solved
// along THIS node's own fixed ray. With u the node's unit direction from the scene centre
// (fixed by its own iTheta/iPhi), c the partner's centre relative to the scene centre, and
// candidate radius r = iR*stepR, the squared distance to the partner is
//
//	|r*u - c|^2 = r^2 - 2*r*(u.c) + |c|^2
//
// Solving that quadratic at = D^2 gives r = (u.c) ± sqrt((u.c)^2 - |c|^2 + D^2); the root
// nearer this node's CURRENT radius is kept (so a turn never flings the node to the far
// side of the scene centre), divided by this node's own effective r-step
// (quantOffset.effectiveSteps()) and rounded to the nearest integer iR. A negative
// discriminant means D is unreachable along this fixed ray at all; per the model, the
// position is then left unchanged — same silent no-op posture the coincident-centres case
// already took.
//
// ONE END OF THE PAIR MOVES, AND IT IS ALWAYS Node2. Node1 is the ANCHOR: it reports its
// own tilt index like any pair member, but it never repositions itself for one. Reuses the
// SAME owner-goroutine commit path as a drag (m.commitLocal, bound in build.go to
// MoveDispatch.commitNodeMoveLocal) — the new world centre is derived from the updated
// offset via the SAME forward path (offsetScenePolar + polar2cart + m.geom.SceneCenter)
// deriveCenters uses, not a second copy of that formula — no new position or commit path,
// no worklist, no coordinator (memory/project_lock_propagation_decentralized.md: a node
// writes only itself).
func (m *nodeGeometry) repositionForTiltIndex(thetaIdx int32) {
	if m.selfKind != "Node2" {
		return
	}
	if m.commitLocal == nil || len(m.partnerCenters) != 1 {
		return
	}
	var partnerID string
	var partnerCenter vec3
	for id, c := range m.partnerCenters {
		partnerID, partnerCenter = id, c
	}
	newOffset, ok := solveTiltIndexROffset(m.quantOffset, m.geom.SceneCenter, partnerCenter, thetaIdx, m.selfKind, m.neighborKinds[partnerID])
	if !ok {
		return
	}
	newPolar := offsetScenePolar(newOffset)
	newPos := m.geom.SceneCenter.Add(polar2cart(newPolar))
	m.commitLocal(m.id, newPos)
}

// solveTiltIndexROffset is the pure helper behind repositionForTiltIndex: given this
// node's CURRENT quantized offset (its fixed iTheta/iPhi ray plus its own effective step
// constants), the scene centre, the partner's world centre, the reported tilt index, and
// both kinds' torus-step counts, returns the offset with ONLY iR changed to the integer
// nearest the radius that puts this node at distance D from the partner along its own
// fixed ray — see repositionForTiltIndex's doc comment for the derivation. ok is false when
// D is unreachable along this ray (negative discriminant); the input offset is returned
// unchanged in that case, but callers must check ok rather than using it.
func solveTiltIndexROffset(offset quantizedOffset, sceneCenter, partnerCenter vec3, thetaIdx int32, selfKind, partnerKind string) (quantizedOffset, bool) {
	idx := thetaIdx
	if idx < 0 {
		idx = -idx
	}
	steps := int(idx) + nodeTorusSteps(selfKind) + nodeTorusSteps(partnerKind)
	D := float64(steps) * wire.BeadStepR

	_, _, rStep := offset.effectiveSteps()
	// u: this node's own unit direction from the scene centre, fixed by its own
	// (unchanged) iTheta/iPhi — radius 1 along the same ray offsetScenePolar would place
	// this node's actual radius on.
	rayPolar := offsetScenePolar(offset)
	rayPolar.R = 1
	u := polar2cart(rayPolar)
	c := partnerCenter.Sub(sceneCenter)
	uDotC := u.Dot(c)
	disc := uDotC*uDotC - c.Dot(c) + D*D
	if disc < 0 {
		return offset, false
	}
	sq := math.Sqrt(disc)
	r1 := uDotC + sq
	r2 := uDotC - sq
	currentR := float64(offset.iR) * rStep
	chosen := r1
	if math.Abs(r2-currentR) < math.Abs(r1-currentR) {
		chosen = r2
	}
	offset.iR = int(math.Round(chosen / rStep))
	return offset, true
}

// applyCenter is the SOLE WRITE of this node's center/reach. It is called ONLY from this
// node's own driving goroutine (handle's moveMsgKindCenter case, driven by fanCenters
// below), which is what makes that one goroutine the exclusive writer of m.geom. It sets
// the held polar position, pushes the fresh center to the dispatch goroutine's owned
// center mirror (m.centerOut, latest-wins — see its doc comment) and to every direct
// neighbor's partnerCenters map (below), and re-emits this node's live geometry.
func (m *nodeGeometry) applyCenter(center vec3, reach float64) {
	setNodeWorld(&m.geom, center)
	m.geom.ReachR = reach
	// Latest-wins non-blocking push onto centerOut: drain any stale unread value first
	// so the slot always ends up holding the newest center, never blocking this
	// goroutine even if the dispatch goroutine hasn't drained the previous push yet.
	select {
	case <-m.centerOut:
	default:
	}
	select {
	case m.centerOut <- center:
	default:
	}
	// Push this fresh center to every direct neighbor (nm.neighborIn's key set — one
	// hop, no cascade) so each neighbor's OWN partnerCenters map picks it up via
	// moveMsgKindNeighborCenter (handle, below). Routed through m.sendMove (this
	// node's own retry queue), same as every other fan-out this file makes, so a
	// momentarily-full neighbor inbox is retried, never dropped or blocking. Sent
	// BEFORE this same commit's broadcastToEdgesAndPartners nil-Center re-emit (called
	// right after applyCenter by every live caller), so per-destination FIFO delivers
	// this push first and the re-emit always sees the just-pushed center.
	for neighborID := range m.neighborIn {
		m.sendMove(neighborID, moveMsg{Kind: moveMsgKindNeighborCenter, NodeID: neighborID,
			SenderID: m.id, FromCenter: center})
	}
	if m.tr != nil {
		m.emitGeometry()
	}
}

// emitGeometry re-emits this node's authoritative geometry (center, radius, ring
// normals — no port geometry: a port carries none, docs/channels-not-ports.md).
// This method and applyCenter both run on this node's own driving goroutine only, so a
// plain field read here can never race a concurrent writer.
func (m *nodeGeometry) emitGeometry() {
	// Dedicated per-node stream (see streamOut's doc comment): write this node's own
	// combined frame immediately on a geometry change, in addition to the tick-driven
	// write in the driving loop's own per-cycle write. NodeGeometry rides THIS frame's
	// own EVENTS section (fully decentralized — it never rides the VIEW stream's
	// fallback bucket) — this node is the sole owner of its own geometry, so it resolves
	// its own NodeRow at the call site (owner_events.go) rather than routing through a
	// shared accumulator.
	m.writeStreamFrame([]wire.RowEvent{{
		Kind: T.KindNodeGeometry, NodeRow: m.nodeRow,
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1,
	}})
}

// writeStreamFrame packs and writes this node's combined per-fd frame (center/radius/
// ring-normals + ports + label + selection-UI columns) to its OWN dedicated fd
// (streamOut). No-op when streamOut is nil (the fallback — see its doc comment) or
// buildFrame was never injected (bare test construction). Called only by this node's own
// driving goroutine, reading m.geom. events carries whatever this call's caller wants
// riding this frame's trailing EVENTS section (nil from a plain tick-driven write).
func (m *nodeGeometry) writeStreamFrame(events []wire.RowEvent) {
	if !m.streamOut.Ok() || m.buildFrame == nil {
		return
	}
	// INVARIANT: a node carries only its OWN events on its OWN dedicated stream. This is
	// the per-goroutine bridge stated in CLAUDE.md's "Bridge surface" and in
	// memory/feedback_no_single_writer_bridge.md + memory/feedback_per_goroutine_bridge.md,
	// and until now it was enforced by prose alone. NodeRow is the ownership column; a
	// FOREIGN node is referenced through TargetRow (see quantized_move.go's abc-drag
	// breadcrumb, which sets NodeRow: nm.nodeRow and TargetRow: the other node). Violating
	// it produces a frame the TS side decodes onto the wrong row — a silently wrong scene
	// that still renders, which is the expensive failure this panic converts into a cheap
	// one. Placed AFTER the nil guard on purpose: bare geometries built in tests never
	// reach the pack path, and nodeRow is seeded alongside streamOut (stream_wiring.go),
	// so any frame that gets here has a real row.
	for _, e := range events {
		if e.NodeRow != m.nodeRow {
			panic(fmt.Sprintf(
				"nodeGeometry.writeStreamFrame: node %q (row %d) is carrying a %s event for row %d on its OWN dedicated stream — NodeRow is an ownership claim, not a reference; a foreign node belongs in TargetRow",
				m.id, m.nodeRow, e.Kind, e.NodeRow))
		}
	}
	center := nodeWorldPos(m.geom)
	sphereR := effectiveRadius(m.geom)
	// This node's own local-frame pole: its own scene-polar direction reversed, so the frame
	// points back at the scene centre (Buffer/layout.go PoleTheta/PolePhi). Derived here from
	// m.geom.ScenePolar — this node's own coordinate, on this node's own goroutine, no
	// neighbour read. Before HasPos there is no direction yet, so the frame stays world +y.
	var poleTheta, polePhi float64
	if m.geom.HasPos {
		poleTheta, polePhi = inwardPole(m.geom.ScenePolar)
	}
	// The DRAWN ring's axis, separate from the navigation pole above (Buffer/layout.go's
	// RingAxisTheta/RingAxisPhi). Default is the torus's own +Z normal, which draws exactly
	// as an unrotated ring did — so a scene that has not asked for anything looks unchanged.
	ringAxisTheta, ringAxisPhi := torusDefaultAxisAngles()
	// topTiltVectorLen is this node's own drawn vector, along the SAME axis as its ring, and 0
	// where a scene draws none (Buffer/layout.go's TopTiltVectorLen). It runs from the node's
	// centre to its own top, so its length IS the node's radius.
	var topTiltVectorLen float64
	if m.upAxis && m.geom.HasPos && len(m.partnerCenters) == 1 {
		// UPRIGHT: the ring STANDS UP along its edge — its plane holds both the edge and
		// world +y, so the node's own up-vector lies IN the ring's plane rather than
		// sticking out of a flat disc. An axis of +y itself would lie the ring flat and
		// put the vector perpendicular to it, which is the opposite arrangement.
		for _, partner := range m.partnerCenters {
			if t, p, ok := uprightRingAxis(nodeWorldPos(m.geom), partner); ok {
				ringAxisTheta, ringAxisPhi = t, p
			}
		}
		topTiltVectorLen = nodeRadius(m.geom.Kind)
	} else if m.coplanarEdges && m.geom.HasPos && len(m.partnerCenters) == 1 {
		// COPLANAR EDGES: swing the axis off the inward pole by the smallest amount that
		// puts the edge INSIDE the ring plane — the inward pole with its along-the-edge
		// component removed. The chain, this node's torus and the beads' own tori then
		// share one plane instead of the chain running through the holes. Only for a node
		// with exactly ONE neighbour: two non-collinear edges have no common plane.
		for _, partner := range m.partnerCenters {
			if t, p, ok := poleContainingEdge(poleTheta, polePhi, nodeWorldPos(m.geom), partner); ok {
				ringAxisTheta, ringAxisPhi = t, p
			}
		}
	}
	// topTiltVectorTheta/topTiltVectorPhi are this node's OWN vector direction — separate from the ring
	// axis above, so a scene/user can aim a node's vector somewhere other than its ring.
	// Never a free float: index × TiltVectorAngleStep (see the constant's own doc comment),
	// the streamed value is pure arithmetic on the integer state this node's own mover
	// holds and persists (m.topTiltVectorThetaIdx/topTiltVectorPhiIdx).
	topTiltVectorTheta := float64(m.topTiltVectorThetaIdx) * CurveParamTiltVectorAngleStep
	topTiltVectorPhi := float64(m.topTiltVectorPhiIdx) * CurveParamTiltVectorAngleStep
	// The BOTTOM TILT VECTOR: streamed straight from this node's own bottomThetaIdx/
	// bottomPhiIdx, decided by THIS node's OWN goroutine (a half turn in θ from its own top
	// tilt index, sign owned by the kind — Node1/Node2's bottomTilt) and reported one-way
	// via PairNodeSelf.SetTiltIndex alongside the top and the normal. Pure mirror here, same
	// as every other index on this frame: this mover derives none of them.
	bottomTiltVectorTheta := float64(m.bottomThetaIdx) * CurveParamTiltVectorAngleStep
	bottomTiltVectorPhi := float64(m.bottomPhiIdx) * CurveParamTiltVectorAngleStep
	// The COPLANAR NORMAL: streamed straight from this node's own normalThetaIdx/
	// normalPhiIdx, which THIS node's OWN goroutine decided (a fixed ±90° in θ from its
	// own tilt index, sign owned by the kind — Node1/Node2's coplanarNormal) and reported
	// one-way via PairNodeSelf.SetTiltIndex. This mover is a pure mirror here, same shape
	// as topTiltVectorTheta/topTiltVectorPhi above — it derives nothing from the edge/partner.
	// Turning the tilt therefore visibly turns the drawn normal WITH it, staying 90° away,
	// instead of the normal staying fixed toward the partner while the tilt moves under it.
	coplanarNormalTheta := float64(m.normalThetaIdx) * CurveParamTiltVectorAngleStep
	coplanarNormalPhi := float64(m.normalPhiIdx) * CurveParamTiltVectorAngleStep
	// The THIRD vector: the direction last received on this node's tilt-vector channel
	// (receivedVectorThetaIdx/PhiIdx, mirrored one-way from this node's own goroutine —
	// see the field's own doc comment). Same length-says-whether-and-how-far convention
	// as topTiltVectorLen: zero when nothing has been received yet (or a reset cleared it),
	// non-zero (this node's own radius, same as topTiltVectorLen) otherwise — so a node with
	// nothing received is distinguishable from one whose received direction happens to be
	// (0,0), which still streams a non-zero length.
	var receivedVectorLen float64
	var receivedVectorTheta, receivedVectorPhi float64
	if m.receivedVectorSet {
		receivedVectorLen = nodeRadius(m.geom.Kind)
		receivedVectorTheta = float64(m.receivedVectorThetaIdx) * CurveParamTiltVectorAngleStep
		receivedVectorPhi = float64(m.receivedVectorPhiIdx) * CurveParamTiltVectorAngleStep
	}
	label := m.geom.Label
	if label == "" {
		label = m.id
	}
	selected, hovered, latchedSel, kindID := m.selected, m.hovered, m.latchedSel, m.kindID
	// This node's own placeholder chain beads, node-local (chain_beads.go). Computed here
	// on this node's own goroutine from its own center + its own partnerCenters map — no
	// cross-goroutine position read.
	chainOX, chainOY, chainOZ, chainLit, chainLitVal, chainBreadcrumbs := m.chainBeads()
	if len(chainBreadcrumbs) > 0 {
		// DIAGNOSTIC ONLY (task/log-node4-chain-aim): chainBeads' own "chain-aim" events,
		// appended here rather than sent via a nested writeStreamFrame call from inside
		// chainBeads (which would recurse back into chainBeads — see that function's doc
		// comment on its breadcrumbs return value).
		events = append(events, chainBreadcrumbs...)
	}
	// nodeID is this node's own numeric identity: ROW ID = NODE ID - 1 (enforced at load,
	// persistence-ownership.md), so it is m.nodeRow+1 by construction — not re-derived by any
	// offline rule the decoder also has to apply, it travels with the frame.
	frame := m.buildFrame(uint32(m.clk.Tick()), m.nodeRow, m.nodeRow+1,
		float32(center.X), float32(center.Y), float32(center.Z),
		float32(nodeRadius(m.geom.Kind)), float32(sphereR),
		verticalRingNormalX, verticalRingNormalY, verticalRingNormalZ,
		flatRingNormalX, flatRingNormalY, flatRingNormalZ,
		float32(poleTheta), float32(polePhi), float32(ringAxisTheta), float32(ringAxisPhi), float32(topTiltVectorLen),
		float32(topTiltVectorTheta), float32(topTiltVectorPhi),
		float32(bottomTiltVectorTheta), float32(bottomTiltVectorPhi),
		float32(coplanarNormalTheta), float32(coplanarNormalPhi),
		float32(receivedVectorLen), float32(receivedVectorTheta), float32(receivedVectorPhi),
		selected, kindID, hovered, latchedSel,
		label, chainOX, chainOY, chainOZ, chainLit, chainLitVal, events)
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(frame)))
	// Fire-and-forget, same reasoning throughout this bridge: no delivery
	// guarantee on this channel, errors ignored.
	_, _ = m.streamOut.Write(hdr[:])
	_, _ = m.streamOut.Write(frame)
}

// flushPending retries every message in m.pending in order, attempting a non-blocking
// send to its destination's inbox. A destination whose channel is momentarily full
// stays in the queue (retried next call) — and so does every LATER item addressed to
// that SAME destination, even if its own channel isn't full, so per-destination FIFO
// is preserved (a retained item is never overtaken by a newer one to the same
// destination). An item whose destination doesn't resolve (unknown id) is dropped,
// matching the old deliverMove no-op for an unknown id. Called only from m's own
// driving goroutine (sendMove, at enqueue time, and the driving loop, every cycle).
func (m *nodeGeometry) flushPending() {
	if len(m.pending) == 0 || m.resolveDest == nil {
		return
	}
	blocked := map[string]bool{}
	kept := m.pending[:0]
	for _, item := range m.pending {
		if blocked[item.destID] {
			kept = append(kept, item)
			continue
		}
		ch, ok := m.resolveDest(item.destID)
		if !ok {
			continue
		}
		select {
		case ch <- item.msg:
		default:
			blocked[item.destID] = true
			kept = append(kept, item)
		}
	}
	m.pending = kept
}
