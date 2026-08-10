// node_geometry_parts.go — the owner sub-structs a nodeGeometry composes.
//
// nodeGeometry used to be a 46-field god-object: its channels, its clock copy, its
// dedicated stream, its selection UI bytes, its tilt mirror indices, its pair readout
// counters, its outgoing wire arrays, its neighbour topology tables and its scene flags
// all sat flat in one namespace, so nothing said which concern a given field belonged to
// and a new field had nowhere to land except "one more loose field". This file gives each
// concern a NAMED type; node_geometry.go keeps the composer plus every method.
//
// Same pattern MoveDispatch already follows (move_dispatch.go, guarded by
// tools/network/structure/check-composer-fields.sh): NAMED sub-objects accessed explicitly (m.ui.selected),
// never Go embedding — embedding would keep the flat namespace and hide the owner.
//
// This is a pure regrouping. Every field keeps its name, its type, its zero value and its
// doc comment; the single-writer-per-nodeGeometry invariant those comments state is
// unchanged — one goroutine still owns the whole composite.
package Wiring

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// nodeMessaging owns this node's own inbound channel set, its outbound retry queue, and
// the closures it routes through — everything about how this node's own goroutine
// receives and hands off a movemsg.Msg. No shared inbox appears anywhere in it.
type nodeMessaging struct {
	// extIn is this node's dedicated channel for EXTERNAL entries — the stdin/gesture
	// goroutine's drag/dragStart sends (md.sendMove). Nothing else ever writes here: no
	// other node shares it.
	extIn chan movemsg.Msg
	// neighborIn holds one dedicated inbound channel PER ADJACENT NODE (keyed by that
	// neighbor's id) — the "two channels, A→B and B→A" topology generalized to this
	// node's whole neighbor set. Built once at construction (newMoveDispatch) from edge
	// adjacency and never mutated afterward, so it's safe for the driving goroutine to
	// snapshot into a fixed select-case list at its own start. A neighbor M's own
	// goroutine is the only writer of neighborIn[M]; nothing else ever sends on it.
	neighborIn map[string]chan movemsg.Msg
	// centerOut is this node's OWN dedicated one-slot delivery channel to the DISPATCH
	// goroutine's owned center mirror (moverRegistry.centerMirror). A size-1 buffered
	// channel written with LATEST-WINS semantics (applyCenter drains any stale unread
	// value before sending the fresh one, never blocking): only the newest pushed center
	// matters to a framing read, so an unread stale value is simply overwritten rather
	// than queued. Only this node's own driving goroutine (applyCenter) ever sends here;
	// only the dispatch goroutine (moverRegistry.drainCenterMirror) ever receives.
	centerOut chan vec3
	// sendMove routes a movemsg.Msg to another id's OWN dedicated channel (resolveDest,
	// below) — no shared inbox, no shared mutable state. Bound to md.enqueueFor(this): it
	// appends to pending and immediately attempts a non-blocking flush (never blocks the
	// calling handler goroutine).
	sendMove func(id string, msg movemsg.Msg)
	// resolveDest looks up the ONE dedicated directed channel FROM this node TO the given
	// destination id — the destination's neighborIn[this node's id] if destID is another
	// node, or the destination edge's srcIn/dstIn depending on which endpoint this node
	// is. There is no shared inbox to look up. nil only in tests that build a bare
	// nodeGeometry directly, in which case flushPending is a no-op.
	resolveDest func(id string) (chan movemsg.Msg, bool)
	// centerOf resolves another node's current world center, bound to md.centerOfNode.
	// Unused by any live handler now that the rule/gate/anchor cascade (which used it to
	// read rule-neighbor centers) is gone; kept wired for any future direct-neighbor
	// lookup need.
	centerOf func(id string) (vec3, bool)
	// commitLocal is the OWNER-GOROUTINE commit path, bound to md.commitNodeMoveLocal
	// (generalized to every node). It publishes this node's own snap SYNCHRONOUSLY via
	// applyCenter instead of enqueuing an async self-send, so it is safe to call from
	// THIS node's own handle() for a movemsg.KindDrag, with no cross-goroutine self-send
	// and no shared mutable state (each node's quantized offset lives on its own
	// geometry — see quantOffset). nil in tests that build a bare nodeGeometry directly.
	commitLocal func(id string, newPos vec3)
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
	tap func(destID string, msg movemsg.Msg)
}

// nodeClocks owns this node's clock pair: the source it copies from once, and its own
// per-goroutine copy. Nothing else in the geometry reads time.
type nodeClocks struct {
	// clockSrc is the Clock this node's driving goroutine Copies from EXACTLY ONCE, at its
	// own start, into clk below (per-goroutine-clock.md). Set once at construction. Not
	// read again after that copy.
	clockSrc clock.Clock
	// clk is this node's OWN clock copy — read by writeStreamFrame (the frame tick) and,
	// for a ring node, by its owning nodeMover's pacing loop (ApplySpeedNonBlocking/
	// SleepCycle). Only the one goroutine driving this geometry ever reads or writes it.
	// Defaults to a fresh, real, live-ticking RealClock (see newNodeGeometry) so a test
	// that never launches a driving goroutine (e.g. a bare literal calling flushPending
	// directly) never dereferences a nil Clock.
	clk clock.Clock
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
}

// NodeFrameInput is one node's own frame values, handed BY NAME to the injected packer
// (NodeFrameBuilder → Buffer.BuildNodeStreamFrame, mapped field-by-name in main.go so this
// package stays Buffer-independent).
//
// WHY A STRUCT: this was a ~35-parameter positional argument list — 20 consecutive
// float32s, then 5 uint8s, then 2 int32s — where transposing two adjacent same-typed
// arguments COMPILED CLEANLY and streamed a silently wrong scene. Named fields make that
// bug class unrepresentable (memory/feedback_make_bug_class_unrepresentable.md).
//
// Field order MIRRORS the packed column order for readability; the wire format is defined
// by Buffer's packing walk, never by this struct's order.
type NodeFrameInput struct {
	// Tick is this node's own clock tick, stamped on the frame.
	Tick uint32
	// NodeRow is this node's stable buffer row; NodeID is its numeric identity
	// (NodeRow+1 by construction — ROW ID = NODE ID - 1, persistence-ownership.md).
	NodeRow int32
	NodeID  int32
	// CX/CY/CZ: this node's own world-space centre.
	CX, CY, CZ float32
	// Radius: this node's own drawn radius. SphereR: the scene sphere's radius.
	Radius, SphereR float32
	// The vertical- and flat-ring normals.
	VRX, VRY, VRZ float32
	FRX, FRY, FRZ float32
	// PoleTheta/PolePhi: this node's own local-frame pole — its scene-polar direction
	// reversed, so the frame points back at the scene centre (world +y before HasPos).
	PoleTheta, PolePhi float32
	// RingAxisTheta/RingAxisPhi: the DRAWN ring's axis, separate from the navigation pole.
	RingAxisTheta, RingAxisPhi float32
	// TopTiltVectorLen encodes WHETHER-AND-HOW-FAR: 0 where the scene draws no vector,
	// otherwise the node's own radius (the vector runs centre → top).
	TopTiltVectorLen float32
	// TopTiltVectorTheta: the vector's OWN direction, separate from the ring axis. Never a
	// free float — index × this node's own lattice step. θ-only, there is no φ.
	TopTiltVectorTheta float32
	// BottomTiltVectorTheta: a half turn in θ from the top index, mirrored straight from
	// this node's own bottomThetaIdx; it shares the top's length column.
	BottomTiltVectorTheta float32
	// CoplanarNormalTheta: a fixed quarter turn in θ from the top index, so turning the
	// tilt visibly turns the normal WITH it.
	CoplanarNormalTheta float32
	// ReceivedVectorLen: same whether-and-how-far convention as TopTiltVectorLen — ZERO
	// when nothing has been received on this node's tilt-vector channel, so a node with
	// nothing received stays distinguishable from one whose received direction is 0.
	ReceivedVectorLen float32
	// ReceivedVectorTheta: that received direction (meaningless at zero length).
	ReceivedVectorTheta float32
	// This node's OWN selection-UI bits and its static NODE_DEFS kind index.
	Selected, KindID, Hovered, LatchedSel uint8
	// LatticePoints: this node's own lattice size — the N the θ columns were converted
	// against.
	LatticePoints uint8
	// RoundsToParallel/MsgsToParallel: this node's own rounds- and messages-to-rest counts.
	RoundsToParallel, MsgsToParallel int32
	// Label: this node's own label bytes, packed inline in its own frame.
	Label string
	// The PARALLEL chain-bead slices (one entry per placeholder bead, node-local offsets).
	ChainBeadOX, ChainBeadOY, ChainBeadOZ []float32
	ChainBeadLit                          []uint8
	ChainBeadLitValue                     []int32
	// Events rides this frame's trailing EVENTS section (nil from a plain tick-driven
	// write). Each must claim this node's OWN row — writeStreamFrame panics otherwise.
	Events []wire.RowEvent
}

// NodeFrameBuilder is the injected node-frame packer (Buffer.BuildNodeStreamFrame behind a
// Buffer-independent seam — see NodeFrameInput). One named-struct argument, not a
// positional run.
type NodeFrameBuilder func(f NodeFrameInput) []byte

// nodeStream owns this node's DEDICATED outbound content-buffer stream — the fd it writes,
// the row that fd's frames claim, the numeric kind column it stamps, and the packer it
// calls (memory/feedback_no_single_writer_bridge.md).
type nodeStream struct {
	streamOut claimedStream
	nodeRow   int32
	kindID    uint8

	buildFrame NodeFrameBuilder
}

// nodeUI owns this node's OWN selection/hover state — per-owner bytes streamed on its own
// frame, never a shared or republished selection map.
type nodeUI struct {
	selected, hovered, latchedSel uint8
	hoverPort                     string
	hoverIsInput                  bool
}

// nodeTilt owns this node's tilt/received vector MIRROR state: integer lattice indices its
// own goroutine (or its kind goroutine, one-way via PairNodeSelf) decided, plus the lattice
// size those indices are converted against. Nothing here is derived from a neighbour.
type nodeTilt struct {
	topTiltVectorThetaIdx  int32
	normalThetaIdx         int32
	bottomThetaIdx         int32
	receivedVectorThetaIdx int32
	receivedVectorSet      bool
	// latticePoints is the point count of THIS node's own lattice — the N a tilt-vector
	// index is converted against (2π / latticePoints per step), reported one-way by
	// PairNodeSelf.SetLatticePoints when a PairNode's own goroutine adopts a new count
	// (PairNode.adoptLattice) and at build time. Defaults to Wiring.FullTurnThetaIdx (the
	// old compile-time constant this field replaces) so every node that never calls
	// SetLatticePoints — every ring node, and a bare test-built pair geometry — streams
	// exactly what it streamed before this field existed.
	latticePoints int32
}

// pairReadout owns the two vector-exchange span counters a pair node's own kind goroutine
// reports one-way for display. Zero on every kind that has no vector exchange.
type pairReadout struct {
	// roundsToParallel is how many vector-exchange rounds this node's own rule took to come
	// to rest after the exchange opened — reported one-way by that node's own goroutine
	// (PairNodeSelf.SetRoundsToParallel), never computed here. Streamed as the Node block's
	// RoundsToParallel column; 0 on every kind that has no vector exchange at all.
	roundsToParallel int32

	// msgsToParallel is the same span counted in vector-channel messages — see the Node
	// block's MsgsToParallel column. Reported alongside roundsToParallel, never derived.
	msgsToParallel int32
}

// nodeOuts owns this node's OUTGOING side: the targets it emits to and the paced wires,
// Outs and step channels those emissions ride. Index-parallel arrays, all written once at
// wiring time.
type nodeOuts struct {
	outTargets     []string
	outWires       []*wire.PacedWire
	outWireTargets []string
	outWireOuts    []*wire.Out
	outStepsIn     []chan int
}

// neighborTopology owns THIS node's own view of who it is adjacent to: incident edge ids,
// each direct neighbour's last-known centre and kind, which of them are mutual, and the
// row lookup used to name a neighbour in this node's own events.
type neighborTopology struct {
	edgeIDs []string
	// partnerCenters is THIS node's OWN copy of every direct neighbor's last-known world
	// center, read by quantized_move.go's neighbor-move math. Written ONLY by this
	// node's own driving goroutine: seeded once at construction (newMoveDispatch,
	// single-threaded setup) from each neighbor's load-time geom, then kept current by
	// the movemsg.KindNeighborCenter handler in handle() below, fed by every direct
	// neighbor's own applyCenter push. Never read or written by any other goroutine.
	partnerCenters map[string]vec3
	neighborKinds  map[string]string
	mutualTargets  map[string]bool
	nodeRowFor     func(id string) (int32, bool)
}

// sceneFlags owns the two scene-wide drawing choices this node applies to its OWN ring
// axis in writeStreamFrame. Read-only after wiring.
type sceneFlags struct {
	coplanarEdges bool
	upAxis        bool
}

// nodeBeads owns this node's placeholder chain-bead actors (bead_chain.go) and the tick
// source they run on.
type nodeBeads struct {
	beadTickFn func() <-chan struct{}
	beadChains map[string]*edgeBeadChain
}
