// Package PairNode is the "PairNode" kind: a pair is two nodes of this one kind. It is
// REACTIVE, not periodic: every cycle it drains its own In and its own
// VectorIn non-blockingly, and runs the straightening rule ITSELF on what arrived. An In
// bead PACES the exchange and decides nothing; the rule lives on the VECTOR channel
// (handleVectorCycle below) — the arrival's distance from the resting state this node is
// holding decides whether it turns and which way, and only if it turned does it reply with a
// vector and place a bead. There are TWO resting states, PERPENDICULAR and PARALLEL, each
// with its own halt, and a node halts on the one it holds and steps over the other. This
// all runs on THIS goroutine: there is no round trip to a second goroutine to decide (see
// TopTiltThetaIdx below for who else the index is reported to and why).
//
// Emission is otherwise silent: with no In arrival there is nothing to react to, and the
// loop is kicked off by a USER — routed here via its own dedicated TiltEditIn channel
// (BuildArgs.TiltEditIn), also drained non-blockingly every cycle. TiltEditIn carries THREE
// distinct edits (task/pair-node-owns-itself split), never conflated:
//
//   - TiltVectorAnglePanel's ▲/▼ click: applies exactly one ±1 step to the named axis, marks
//     this end HELD (a tilt a user set is intent, not error — the partner moves to restore
//     square instead of this end turning on an arrival), and ALSO opens the vector exchange
//     by sending this node's own outgoing vector alongside a bead — a click that only moved an
//     index would leave the partner with nothing to answer.
//   - the START TILT button (TiltVectorButtons.tsx, TiltEditMsg.Start): opens the vector
//     exchange from whatever angles are CURRENTLY set — sends this node's own outgoing
//     vector alongside a bead ("THE KICK"), which is what gives handleVectorCycle something
//     to reply to; a channel whose only sends are replies never carries anything at all. It
//     changes NO index of its own. With both nodes of a pair perpendicular nothing
//     circulates on In, correctly, since there is nothing left to straighten, so the loop
//     has no way to start on its own — Start is the thing a user clicks to start it.
//     Pairing two PairNode instances with one edge each direction (a.Out → b.In,
//     b.Out → a.In) needs no seed/bootstrap node: nothing ever sends until a user
//     starts it, so there is no deadlock to bootstrap out of at t=0.
//   - the RESET button (TiltVectorButtons.tsx, TiltEditMsg.Reset): the opposite of Start — it
//     places NO bead, a stop-and-return, not a nudge, so it never starts the straightening
//     exchange. It does more than zero the indices, because zeroed indices are not by
//     themselves a stopped exchange: it runs this node's full clear() (below), which also
//     empties the bead edge — the thing that has actually been turning these tilts — so
//     nothing is left in the pair that could land a moment later and step it back off zero.
//
// THE KIND'S LOGIC IS THIS FILE: the Node struct, the builder that constructs it, the Update
// loop that IS this node's goroutine, and the two functions that decide — stepFromVector and
// handleVectorCycle. Everything that only SERVES those decisions lives with the concern it
// belongs to: the tilt machine and which mode this node adopts in machine.go, the lattice and
// this node's two ends on it in ring.go, the directions it computes and reports in vectors.go,
// and the panel-driven edits plus what a reset undoes in edits.go.
package PairNode

import (
	"context"
	"fmt"
	"strconv"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

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

type Node struct {
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

	// Ring is THIS NODE'S OWN lattice — its states, and the counts every rule reads off them.
	// The point count is a scene setting a user can change, so this is not fixed for the life
	// of the process; a change means this goroutine building itself a new ring, never a
	// shared one being rewritten under other readers.
	//
	// nil means the default lattice, which is what a bare test build gets — see ringOf below,
	// the one read of this field.
	Ring *ring
	// TiltEditIn is this node's dedicated channel for a panel-driven tilt-angle click
	// (TiltVectorAnglePanel), claimed at build time via BuildArgs.TiltEditIn — see the
	// package doc comment's "THE KICK".
	TiltEditIn <-chan Wiring.TiltEditMsg
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
	// SyncTiltIndex notifies this node's own geometry of the current TopTiltThetaIdx AND the
	// current coplanar-normal index (coplanarNormal, below) — one-way, fire-and-forget,
	// never an ack (BuildArgs.SyncTiltIndex).
	SyncTiltIndex func(theta, normalTheta, bottomTheta int32)
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

	// rest is this node's own rest instrumentation — see restCounters. Named, not embedded,
	// so every read says which concern it came from (n.rest.roundsSinceOpen).
	rest restCounters
}

func (n *Node) clock() wire.Clock {
	if n.Clock == nil {
		return wire.NewRealClock()
	}
	return n.Clock
}

// stepFromVector decides whether an arrived direction turns this node at all, and if so which
// way. THERE ARE TWO TARGETS, and a node is returning to exactly one of them:
//
//		PERPENDICULAR   angle length 0 or a half turn   the two tilts are a quarter turn apart
//		PARALLEL        angle length a quarter turn     the two tilts on one line, either way round
//
//	  - not holding either yet -> the first halt reached is taken up, and the node stops there.
//	  - holding one, arrival IS it -> stand still. Nothing else writes what is held, so an
//	    arrival part-way back leaves the node still knowing where it is returning to.
//	  - otherwise -> ONE step, the way that leaves this node nearer its own halt (stepToward).
//
// The other halt is stepped straight over. It has to be: the two sit a quarter turn apart in
// angle length, so the walk back to one crosses the other, and a node that stopped at any halt
// was taken by whichever it touched first — the log showed a pair holding perpendicular walk
// correctly toward angle length 0, meet 12 on the way, and take up parallel there, in both
// directions of disturbance.
//
// Both ends of a pair run this same unmodified rule, and both directions of travel are links
// rather than ±1, so a step cannot leave the ring. The pairing that matters is with what
// outgoingVector sends: this reads an arrival that is the partner's coplanar normal as-is.
// Worked run: docs/pair-node/vectors.html.
func (n *Node) stepFromVector(received Wiring.TiltVectorMsg) bool {
	arrival := n.ringOf().arrivedState(received.ThetaIdx)
	before := n.topState()

	// A NODE HALTS ON ITS OWN HALT AND STEPS THROUGH THE OTHER ONE. Until it is holding either,
	// the first halt it reaches is the one it takes up; after that, only that halt stops it.
	// Landing on the other is nothing — an angle it happens to be passing over on the way back
	// to its own, which is what capturing it there got wrong.
	//
	// NOTHING HERE CHOOSES A MACHINE. Which one this node runs was decided outside it, from
	// the gap, at the moment of the click that set the tilt (applyTiltEdit above,
	// nodes/Wiring/tilt_machine_chooser.go). An arrival cannot answer that question: a node
	// with nothing to return to closes on the arrival, and closing on the arrival is the
	// perpendicular measure, so inferring it here always answered perpendicular and a pair set
	// up near perpendicular could never be asked to run the parallel machine.
	//
	// A node still in the setting mode — before any click, or after a reset — moves nothing, and
	// needs no test here to make that happen: every angle length is that mode's home, so it is
	// already halted wherever it stands and the step below is not reached.
	// THE END THAT WAS MEASURED IS THE END THAT MOVES. step names it; setTop/setBottom write it
	// and derive the other. Behaviour is the same either way — the ends are a half turn apart, so
	// a slot gained by one is a slot gained by both — but the write now says which end the
	// arrival was near, instead of expressing a bottom-side turn as a negated top-side one.
	if !n.Machine.settled(before, arrival) {
		if moved, atBottom := n.Machine.step(before, arrival); atBottom {
			n.setBottom(moved)
		} else {
			n.setTop(moved)
		}
	} else {
		// Came to rest on this arrival. The freeze happens in handleVectorCycle, AFTER the
		// reply goes out, so the message that closes the round is inside the count.
		n.rest.restedThisCycle = true
	}
	return true
}

// handleVectorCycle is PairNode's WHOLE per-cycle vector-channel loop body: read
// VectorIn non-blocking; if something arrived, step (stepFromVector decides whether this node
// turns at all, and which way — see its own doc comment for the one target, square); and if it
// stepped, send outgoingVector back out on VectorOut, also non-blocking, and place the paired
// bead. On a false return from stepFromVector (the arrival landed exactly on this node's own
// top) this sends nothing and places no bead — that is the ONE way the vector exchange comes to
// rest. A RESET marker (below) is the other way the exchange stops. This never touches In/Out
// or beads on the halt path — the vector channel is a separate, additive exchange.
func (n *Node) handleVectorCycle(tick int64) {
	received, ok := Wiring.PollRecvVector(n.VectorIn)
	if !ok {
		return
	}
	n.rest.msgsSinceOpen++ // this node's own receive
	n.rest.roundsSinceOpen++
	// A RESET marker is not a direction to act on: run this node's FULL clear (indices,
	// third arrow, vector end, delivered beads, and the beads still crossing this node's
	// own outgoing wires) and REPLY WITH NOTHING. Replying would bounce the reset back and
	// forth forever; the marker's job is to stop the exchange, so it ends here. This is
	// the clear that actually makes the pair quiescent — see clear's own doc comment on
	// why the marker-driven one, not the button-driven one, is the one that lands last.
	if received.Reset {
		n.clear()
		return
	}
	// EVERY VECTOR MESSAGE SAYS WHICH MACHINE ITS SENDER IS RUNNING (outgoingVector), so this
	// is how the end that did not decide learns the choice: from the first reply. Adopting
	// STICKS, so a later message cannot switch a running machine — see adoptMachine.
	n.adoptMachine(received.Machine)
	// A DIRECTION FROM ANOTHER LATTICE IS NOT A DIRECTION HERE. The two ends of a pair adopt
	// a new point count at their own moments, each on its own goroutine, so between those
	// moments an index picked on the old lattice can land here — where it names a different
	// angle, or no state at all. Dropping it is the definite answer: the partner adopts the
	// same count within its own next cycle and the exchange resumes from directions both
	// ends can read. Zero is a bare test build that stated nothing, and is taken as this
	// node's own lattice.
	if received.Points != 0 && received.Points != n.ringOf().points {
		return
	}
	// A real direction. It is recorded UNCONDITIONALLY — before, and independently of, the
	// step decision below — and then STAYS until the next arrival replaces it. It does not
	// vanish when the exchange settles: the last direction a node was sent is what it is
	// still holding, and blanking the arrow the moment the pair stops turning would erase
	// the very state the pair came to rest in. The only thing that removes it is a RESET
	// (clear, above — this node's own or its partner's marker), which removes it because a
	// reset means there is nothing in the pair at all any more.
	n.ReceivedThetaIdx = received.ThetaIdx
	n.ReceivedSet = true
	n.syncReceivedVector()
	// THE GAP IS CHECKED WHEN THE EXCHANGE OPENS, which is the first arrival — START is the
	// moment the setup is finished, and it is also the first moment either end can see BOTH
	// tilts. The arrival is the partner's normal, a quarter turn off its tilt, so backing that
	// quarter out gives the partner's own tilt and the gap between the two is a real
	// measurement rather than one node's angle against zero.
	//
	//	the gap is a quarter turn  ->  perpendicular machine
	//	anything else (acute)      ->  parallel machine
	//
	// Only while still in the setting mode: this is the setup being read, and after that a click
	// is a jitter for the running machine to correct, not a new instruction. The other end learns
	// the answer from this node's next reply, which carries it (outgoingVector).
	if n.Machine == setting {
		n.adoptMachine(n.machineForGap(n.ringOf().arrivedState(received.ThetaIdx)))
	}
	if !n.stepFromVector(received) {
		return
	}
	n.syncTiltIndex()
	n.rest.msgsSinceOpen++ // this node's own reply
	Wiring.SendVectorLatestNonBlocking(n.VectorOut, n.outgoingVector())
	if !n.rest.restReported {
		// LIVE while the tilt is still turning: report the running counts each round so the
		// readout climbs as it goes rather than staying blank and then jumping. Once this
		// node comes to rest the same numbers are reported one last time and then frozen —
		// the exchange keeps circulating after rest (this function replies to every arrival
		// whether or not it moved), so a counter that kept reporting would measure how long
		// the scene had been open instead of how far the tilt travelled.
		n.rest.roundsAtRest = n.rest.roundsSinceOpen
		n.rest.msgsAtRest = n.rest.msgsSinceOpen
		if n.Self != nil {
			n.Self.SetRoundsToParallel(n.rest.roundsAtRest, n.rest.msgsAtRest)
		}
		if n.rest.restedThisCycle {
			n.rest.restReported = true
		}
	}
	n.rest.restedThisCycle = false
	// The bead rides along with the vector: one message, one visible bead, so the bead
	// loop ends exactly when the exchange does. THIS goroutine is still the sole placer on
	// this Out (wire.Out.PlaceDrivenAt's one-goroutine-per-Out invariant) — the placement
	// only moved between two branches of this same loop.
	if n.Out != nil {
		n.Out.PlaceDrivenAt(1, tick)
	}
}

func (n *Node) Update(ctx context.Context) {
	wire.TryEmit(n.EmitGeometry)
	// This node's own mover-owned startup geometry emit — see Self's own doc comment.
	// There is no separate nodeMover goroutine to make this emit any more.
	n.Self.EmitGeometryOnce()

	// Report THIS node's OPENING tilt/normal pair once, before the loop. Self is a
	// passive mirror of these (PairNodeSelf.SetTiltIndex) and has no way to derive the
	// normal itself, so without this its normal indices sit at their zero value until the
	// first arrival or panel click — and since the tilt index opens at 0 too, both
	// directions decode to world +y and the two drawn arrows superimpose, which reads as
	// the coplanar normal being missing entirely.
	n.syncTiltIndex()

	// Copy taken ONCE at this goroutine's start (Update IS the goroutine).
	clk := n.clock().Copy()

	for {
		if ctx.Err() != nil {
			return
		}

		// Drain In non-blocking. A bead arrival PACES the exchange and marks the round
		// trip; it DECIDES nothing. It used to step this node's tilt one click in this
		// kind's own fixed direction, with no reference to anything that arrived — so
		// every bead round trip turned this node the same way forever, independently of
		// (and on top of) the acute-test rule that is supposed to own that decision. Two
		// rules moved one index: when they agreed the node double-stepped, when they
		// disagreed they cancelled and it froze. The tests are now the only thing that turns a tilt
		// on an arrival, and the bead is what makes that turn visible and timed.
		//
		// It does not place a bead onward either: the bead now travels WITH the vector,
		// placed by handleVectorCycle when the tests actually move this node, so the bead
		// loop lives and dies with the exchange it is pacing instead of circulating on
		// its own.
		if _, ok := n.In.PollRecv(); ok {
			if n.Fire != nil {
				n.Fire()
			}
		}

		// Drain TiltEditIn non-blocking: a panel/RESET/START edit — see the package doc
		// comment for the three-way split. applyTiltEdit decides placeBead: true for Start
		// and for a plain adjust (both open the exchange), false only for Reset.
		if n.TiltEditIn != nil {
			select {
			case edit := <-n.TiltEditIn:
				placeBead := n.applyTiltEdit(edit)
				n.syncTiltIndex()
				if placeBead && n.Out != nil {
					n.Out.PlaceDrivenAt(1, clk.Tick())
				}
			default:
			}
		}

		// Drain LatticeIn non-blocking: a new point count for this node's own ring. Drained
		// BEFORE the vector cycle below so that anything already queued on VectorIn is
		// discarded by the adopt rather than read one last time against the lattice it was
		// not picked on.
		if n.LatticeIn != nil {
			select {
			case points := <-n.LatticeIn:
				n.adoptLattice(points)
			default:
			}
		}

		// Vector-channel exchange: the ONE place an arrival turns this node's tilt, and
		// the place the outgoing bead is now placed from — see handleVectorCycle's own
		// doc comment.
		n.handleVectorCycle(clk.Tick())

		// This node's own mover work — drain its own dedicated inbound channels
		// (drag/select/hover/center/neighborCenter/etc.), drive its own outgoing
		// wires one cycle, retry pending sends, write its own dedicated stream
		// frame. Run on THIS goroutine, on THIS node's own clock tick: there is no
		// separate nodeMover goroutine for this node any more (task/pair-node-owns-
		// itself) — see Self's own doc comment.
		n.Self.Step(ctx, clk.Tick())

		wire.ApplySpeedNonBlocking(clk, n.SpeedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}

func init() {
	// PairNode CONSTRUCTS ITSELF (Wiring.RegisterBuilder), same self-construction
	// shape as every other kind — see Pacer/Input for the general note on why
	// this replaced reflectBuild.
	Wiring.RegisterBuilder("PairNode",
		[]Wiring.PortSpec{
			{Name: "In", Dir: Wiring.PortIn},
			{Name: "Out", Dir: Wiring.PortOut},
		},
		func(a Wiring.BuildArgs) (wire.Node, error) {
			n := &Node{
				Clock: wire.NewRealClock(),
			}
			// This node's own spec id, which is what START is addressed by — see PairID's
			// own doc comment. A name that is not a number leaves PairID at 0, so such a
			// node simply never opens an exchange rather than silently becoming id 1.
			if id, err := strconv.Atoi(a.Name()); err == nil {
				n.PairID = int32(id)
			}
			n.Fire = a.Fire()
			if clk := a.Clock(); clk != nil {
				n.Clock = clk
			}
			n.SpeedCh = a.SpeedCh()
			n.In = a.In("In")
			n.Out = a.Out("Out")
			// The persisted seed is a NUMBER from outside this kind — an old position.json
			// can hold anything, including a running count from before the tilt became a
			// state — so it comes in through seedState, which asks the ring which state
			// carries that index. After this line the tilt is a state and stays one.
			// This node's own lattice, opened at the scene's currently-persisted point
			// count (view/lattice.json via BuildArgs.LatticePointsSeed) rather than the
			// compile-time default.
			latticeSeed := a.LatticePointsSeed()
			n.Ring = newRing(latticeSeed)
			seed, seedUnknown := n.Ring.seedState(a.TiltVectorAngleSeed())
			n.setTop(seed)
			n.TiltEditIn = a.TiltEditIn()
			n.LatticeIn = a.LatticeIn()
			// Self replaces the old SyncTiltIndex/SyncReceivedVector/ClearOutBeads
			// messages-to-a-separate-mover-goroutine (task/pair-node-owns-itself):
			// this node's own goroutine now owns that mover state directly, so what
			// used to be a message is a plain method call on the same object below.
			self := a.ClaimSelfDrive()
			n.Self = self
			n.SyncLatticePoints = func(points int32) {
				self.SetLatticePoints(points)
			}
			n.SyncLatticePoints(latticeSeed)
			if seedUnknown {
				// The persisted index is not one this ring has — a position.json written
				// before the tilt became a state, or by a build with a different lattice.
				// The node opens at the origin and says which number it refused, rather
				// than computing some other direction and drawing it as if chosen.
				self.Breadcrumb("pair-seed-unknown", fmt.Sprintf(
					"node=%s persisted=%d loaded=%d", a.Name(), a.TiltVectorAngleSeed(), seed.idx))
			}
			n.SyncTiltIndex = func(theta, normalTheta, bottomTheta int32) {
				self.SetTiltIndex(theta, normalTheta, bottomTheta)
			}
			n.SyncReceivedVector = func(theta int32, set bool) {
				self.SetReceivedVector(theta, set)
			}
			n.ClearOutBeads = func() { self.ClearOutBeads() }
			n.VectorOut = a.VectorOut()
			n.VectorIn = a.VectorIn()
			// EmitGeometry stays nil deliberately — n.Self.EmitGeometryOnce (Update)
			// makes this node's own startup geometry emit instead.
			return n, nil
		})
}
