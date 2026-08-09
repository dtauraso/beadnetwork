package PairNode

// node_parts.go — the owner sub-structs a PairNode Node composes.
//
// Node used to be a 21-field god-object: the channels and closures its builder plumbed in,
// its two ends on the ring and the machine they run, its lattice, its vector exchange and
// its rest counters all sat flat in one namespace, so nothing said which concern a given
// field belonged to and a new field had nowhere to land except "one more loose field". This
// file gives each concern a NAMED type; node.go keeps the composer, the builder, the Update
// loop and the two functions that decide.
//
// Same pattern nodes/Wiring/node_geometry_parts.go follows (and tools/check-composer-fields.sh
// guards there): NAMED sub-objects accessed explicitly (n.vec.VectorOut), never Go embedding
// — embedding would keep the flat namespace and hide the owner.
//
// This is a pure regrouping. Every field keeps its name, its type, its zero value and its doc
// comment; the single-writer-per-Node invariant those comments state is unchanged — one
// goroutine still owns the whole composite. NO FIELD WAS ADDED: a field added to preserve a
// behaviour is how this kind was broken once before (session-log.md, "One pair implementation,
// and the field that hid it").

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// nodePlumbing owns EVERYTHING THIS NODE'S BUILDER PLUMBED INTO IT and nothing this node
// decides: its own spec id, its bead ports, its clock and speed channel, the trace/geometry
// closures it was handed, and its own geometry object. No rule reads any of it to work out
// where a tilt turns — it is how this node is connected, not what it does.
type nodePlumbing struct {
	// PairID is THIS NODE'S OWN SPEC ID, the number the editor draws on it — the builder
	// parses it from BuildArgs.Name(), which is the node's directory name under topology/
	// and is a number by construction (.claude/rules/persistence-ownership.md: node ids ARE
	// numbers, strings only because they are directory names).
	//
	// It decides ONE thing: START opens the exchange from id 1 alone (applyTiltEdit). The
	// panel posts START to every node row, because the webview holds no domain knowledge
	// about which node should open; Go decides here, by id.
	//
	// It decides NOTHING ELSE. There is one pair kind and one implementation, so both ends
	// of a pair derive the same directions and step the same way, and the id is the only
	// thing that distinguishes them at all. A pair therefore needs the node numbered 1 to be
	// one of its ends — with no id 1 present, nothing opens the exchange and START does
	// nothing.
	//
	// The zero value is what a bare test build in this package constructs; it is not id 1,
	// so such a node does not open an exchange unless the test says PairID: 1.
	PairID int32

	Fire         func()
	EmitGeometry func()
	// Clock is this node's OWN clock storage, assigned by this kind's own
	// builder directly from the loader's origin (per-goroutine-clock.md; see
	// input.Node.Clock for the fuller rationale). Update() Copies it once for
	// its own loop — the sole clock-owning goroutine this node has.
	Clock wire.Clock
	// SpeedCh delivers a speed change to this goroutine's own clk copy.
	// Assigned by this kind's own builder; nil on a test build with no loader.
	SpeedCh <-chan float64
	// In is one of two triggers that drive this node — see the package doc comment.
	In *wire.In
	// Out is the sole output. THIS goroutine is now the SOLE placer on it (below),
	// preserving wire.Out.PlaceDrivenAt's one-goroutine-per-Out invariant — nothing else
	// places on this Out at all.
	Out *wire.Out
	// ClearOutBeads drops every bead still crossing this node's outgoing wires — a call on
	// this node's own Self (BuildArgs.ClearOutBeads), not a message to a second goroutine.
	// Called only from clear(), below: this goroutine drives those wires, so it clears them
	// through the same object that drives them rather than reaching into the wire.
	ClearOutBeads func()
	// Self is this node's own geometry/mover state (task/pair-node-owns-itself,
	// Wiring.PairNodeSelf), claimed at build time via BuildArgs.ClaimSelfDrive. THIS
	// goroutine (Update, below) is the sole driver of it — there is no separate
	// nodeMover goroutine for this node any more. nil on a bare test build with no
	// loader; every PairNodeSelf method is nil-safe.
	Self *Wiring.PairNodeSelf
}

// tiltHeld owns THE TILT THIS NODE IS HOLDING: its two ends on the ring, the machine that
// says where those ends are returning to, the dedicated channel a user's panel click arrives
// on, and the one-way report that mirrors the ends into this node's own geometry. It is the
// state the pair rule reads and writes; nothing else in the struct is.
type tiltHeld struct {
	// Top is THIS node's OWN tilt direction, held as a STATE ON THE RING (ring.go) rather
	// than as a number this goroutine does arithmetic to. Turning is following a link, so
	// there is no index to keep in range and no twenty-fifth direction to land on.
	//
	// THIS GOROUTINE IS THE ONE WRITER, full stop: seeded once at build time from the persisted
	// value (BuildArgs.TiltVectorAngleSeed, mapped through stateFor) and moved ONLY by this
	// goroutine's own Update loop, below.
	//
	// Of the two fields it is not the one writer, and cannot be: an update moves the end the
	// arrival was measured at, which is Bottom for half the ring. The two CANNOT DISAGREE
	// because neither is ever written alone — one step writes the end that moved and reads the
	// other straight off its own opposite link, in the same statement, so there is no window
	// where one has turned and the other has not, and no arithmetic that could put them
	// anywhere but a half turn apart.
	//
	// Every change is reported one-way to this node's own
	// geometry (SyncTiltIndex, i.e. Self) so the geometry — which owns streaming this node's
	// scene columns and persisting them to its own position.json — stays in sync; the
	// geometry never decides or mutates this itself, it mirrors what it is told.
	//
	// nil means the ring's origin: a bare test build constructs this struct without a
	// builder, and topState below reads a nil Top as direction 0 rather than making every
	// such test say so. There is no companion φ — every tilt vector in this exchange lives in
	// the θ-only plane (memory/feedback_abc_times_constant_not_rederive.md: an index times a
	// step constant, trig only at the cartesian/polar boundary).
	Top *tiltState

	// Bottom is the OTHER end of the same line, a half turn from Top, and it is stored rather
	// than fetched through Top.opposite so that an update can NAME the end it drove. The rule
	// counts from whichever end the arrival is nearer (machine.go, nearerEndCount) and moves
	// that one; with only Top held, the half of the ring measured at the bottom had to be
	// written back to the top with its direction negated, which is a correction term standing
	// in for a fact the state could not say.
	//
	// It changes nothing about what is persisted or streamed: position.json still carries one
	// angle, and this end is still a half turn from it. What is stored is which end an update
	// names, not a second degree of freedom.
	//
	// nil means the ring origin's opposite, for the same reason Top's nil means the origin.
	Bottom *tiltState

	// Machine is THIS NODE'S tilt machine — one instance, carrying which mode it is in. The
	// modes differ only in the angle lengths they call home (machine.go). A node has to know
	// which it is in, because that is what says where it is returning to when something
	// disturbs it.
	//
	// Its ZERO VALUE IS THE SETTING MODE, so a Node built as a literal is already in the mode a
	// node starts in. There is no nil and no "no machine" state to test for.
	//
	// It is set when an arrival lands on one machine's halt and by nothing else — no arrival in
	// between erases it, so a node disturbed mid-turn still knows what it is returning to. The
	// RESET button is the one thing that erases it (clear).
	Machine tiltMachine

	// TiltEditIn is this node's dedicated channel for a panel-driven tilt-angle click
	// (TiltVectorAnglePanel), claimed at build time via BuildArgs.TiltEditIn — see the
	// package doc comment's "THE KICK".
	TiltEditIn <-chan Wiring.TiltEditMsg
	// SyncTiltIndex notifies this node's own geometry of the current TopTiltThetaIdx AND the
	// current coplanar-normal index (coplanarNormal, below) — one-way, fire-and-forget,
	// never an ack (BuildArgs.SyncTiltIndex).
	SyncTiltIndex func(theta, normalTheta, bottomTheta int32)
}

// latticeState owns THE LATTICE THIS NODE'S DIRECTIONS ARE INDICES ON: its own ring, the
// dedicated channel a new point count arrives on, and the one-way report of the count it
// adopted. It says how many directions there are, never which one this node holds.
type latticeState struct {
	// Ring is THIS NODE'S OWN lattice — its states, and the counts every rule reads off them.
	// The point count is a scene setting a user can change, so this is not fixed for the life
	// of the process; a change means this goroutine building itself a new ring, never a
	// shared one being rewritten under other readers.
	//
	// nil means the default lattice, which is what a bare test build gets — see ringOf below,
	// the one read of this field.
	Ring *ring
	// LatticeIn carries a new POINT COUNT for this node's own ring — the scene setting the
	// angles panel changes, delivered to every pair node on its own dedicated channel, the
	// same shape as TiltEditIn. Drained non-blocking every cycle; a value that matches the
	// count this node already runs is a no-op (adoptLattice). nil on a bare test build.
	LatticeIn <-chan int32
	// SyncLatticePoints notifies this node's own geometry of the current lattice point
	// count — one-way, fire-and-forget, never an ack, same shape as SyncTiltIndex below.
	// The geometry converts a tilt-vector INDEX to an angle every frame (2π / points per
	// step) but does not itself decide the count, so it has to be told whenever this
	// goroutine adopts a new one (BuildArgs.SyncLatticePoints, i.e. Self.SetLatticePoints).
	SyncLatticePoints func(points int32)
}

// vectorExchange owns THIS NODE'S TWO ENDS OF THE TILT-VECTOR CHANNEL and the one thing it
// keeps from what arrived there — the last received direction, the third drawn arrow, plus
// the one-way report that mirrors it into this node's own geometry. It carries directions;
// it decides nothing about them.
type vectorExchange struct {
	// VectorOut/VectorIn are THIS node's own ends of its dedicated tilt-vector channel
	// (Wiring.TiltVectorMsg — an integer θ index, never floats on a channel),
	// claimed at build time via BuildArgs.VectorOut/VectorIn. It travels ALONGSIDE the
	// ordinary bead edge (In/Out above), never replacing it — beads are unaffected.
	// Buffered depth 1, latest-wins, non-blocking on both ends
	// (Wiring.SendVectorLatestNonBlocking / Wiring.PollRecvVector). nil when this
	// node's edge partner did not also ask for one, or on a bare test build with no
	// loader — both helpers already treat nil as "nothing wired".
	VectorOut chan<- Wiring.TiltVectorMsg
	VectorIn  <-chan Wiring.TiltVectorMsg
	// ReceivedThetaIdx/ReceivedSet are THIS node's own record of the LAST
	// direction that ARRIVED on VectorIn — the third drawn arrow (user request: "show a
	// 3rd vector...the last iteration of it as a different color in the node that
	// received it"). Written ONLY by this goroutine, in handleVectorCycle below: an
	// arrival REPLACES whatever was here before (never accumulates), and it persists
	// indefinitely otherwise — it is NOT cleared when the straightening exchange settles.
	// It IS cleared by a RESET, both this node's own (applyTiltEdit's Reset branch) and a
	// Reset marker arriving on VectorIn (handleVectorCycle's Reset branch): a reset is a
	// stop-and-return, and a stale received arrow left hanging would contradict that.
	// Reported one-way to this node's own geometry via SyncReceivedVector, same shape as
	// TopTiltThetaIdx/SyncTiltIndex above.
	ReceivedThetaIdx int32
	ReceivedSet      bool
	// SyncReceivedVector notifies this node's own geometry of the current
	// ReceivedThetaIdx/Set — one-way, fire-and-forget, never an ack
	// (BuildArgs.SyncReceivedVector).
	SyncReceivedVector func(theta int32, set bool)
}

// restCounters owns this node's REST INSTRUMENTATION and nothing else: how many messages and
// rounds its own vector exchange has run since it opened, the pair of counts frozen at the
// moment its rule first came to rest, and the two flags that make "first" mean first. It is
// pure counting and reporting ABOUT the exchange — no part of the pair rule reads it, and
// nothing in it can change where a tilt turns.
//
// Same pattern node_geometry_parts.go follows (and tools/check-composer-fields.sh guards
// there): a NAMED sub-object accessed explicitly (n.rest.roundsSinceOpen), never Go embedding
// — embedding would keep the flat namespace and hide the owner.
type restCounters struct {
	// msgsSinceOpen counts THIS node's own vector-channel messages since the exchange
	// opened — every receive and every reply. Both directions are counted because both are
	// work this node did: a node that answers two arrivals received twice and replied
	// twice, four messages, and counting only the arrivals would report half of what it
	// moved through.
	//
	// The START opener is NOT counted: it is the kick, not a round, and including it made
	// the opening end read one higher than the other for identical work.
	msgsSinceOpen int32

	// roundsSinceOpen counts the same span in ROUNDS: arrivals this node answered. One
	// answered arrival is one round from this node's side, and is two messages.
	roundsSinceOpen int32

	// roundsAtRest is the count REPORTED to the geometry: live each round while the tilt is
	// still turning, then frozen at the value it held when this node's rule FIRST came to
	// rest after the exchange opened — the number reported to the geometry and streamed as
	// the Node block's RoundsToParallel column.
	//
	// It freezes because the exchange does not stop when the pair settles: stepFromVector
	// replies to every arrival whether or not it moved, so msgsSinceOpen keeps climbing
	// for as long as the scene is open. A reader wants how far the tilt had to travel, and
	// that stops changing at rest.
	//
	// restReported is what makes "first" mean first: without it, every later arrival would
	// re-report the then-current count and the column would climb after all.
	roundsAtRest int32
	msgsAtRest   int32
	restReported bool

	// restedThisCycle is set by stepFromVector when this node's rule found itself already
	// at its halt, and read by handleVectorCycle after that cycle's reply has been sent.
	// It exists so the freeze lands after the send rather than before it — see roundsAtRest.
	restedThisCycle bool
}
