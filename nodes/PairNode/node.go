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
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// Node COMPOSES the five owners its state is split across (node_parts.go) and adds nothing of
// its own — every field of this kind lives in exactly one of them, named, never embedded, so
// every read says which concern it came from (n.tilt.Top, n.vec.VectorIn).
type Node struct {
	// plumb is what this node's builder plumbed in: its id, ports, clock and geometry —
	// see nodePlumbing.
	plumb nodePlumbing
	// tilt is the tilt this node is holding and the machine it runs — see tiltHeld.
	tilt tiltHeld
	// lattice is the ring its directions are indices on — see latticeState.
	lattice latticeState
	// vec is its two ends of the tilt-vector channel and what last arrived — see
	// vectorExchange.
	vec vectorExchange
	// rest is this node's own rest instrumentation — see restCounters. Named, not embedded,
	// so every read says which concern it came from (n.rest.roundsSinceOpen).
	rest restCounters
}

func (n *Node) clock() clock.Clock {
	if n.plumb.Clock == nil {
		return clock.NewRealClock()
	}
	return n.plumb.Clock
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
// Worked run: docs/pair-node/math/vectors.html.
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
	if !n.tilt.Machine.settled(before, arrival) {
		if moved, atBottom := n.tilt.Machine.step(before, arrival); atBottom {
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
	received, ok := Wiring.PollRecvVector(n.vec.VectorIn)
	if !ok {
		return
	}
	n.rest.countArrival()
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
	if n.fromAnotherLattice(received) {
		return
	}
	n.recordReceived(received)
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
	if n.tilt.Machine == setting {
		n.adoptMachine(n.machineForGap(n.ringOf().arrivedState(received.ThetaIdx)))
	}
	if !n.stepFromVector(received) {
		return
	}
	n.reply()
	n.reportRest()
	// The bead rides along with the vector: one message, one visible bead, so the bead
	// loop ends exactly when the exchange does. THIS goroutine is still the sole placer on
	// this Out (wire.Out.PlaceDrivenAt's one-goroutine-per-Out invariant) — the placement
	// only moved between two branches of this same loop.
	if n.plumb.Out != nil {
		n.plumb.Out.PlaceDrivenAt(1, tick)
	}
}

func (n *Node) Update(ctx context.Context) {
	n.openingEmit()

	// Copy taken ONCE at this goroutine's start (Update IS the goroutine).
	clk := n.clock().Copy()

	for {
		if ctx.Err() != nil {
			return
		}

		n.paceOnBeadArrival()
		n.drainTiltEdit(clk)
		n.drainLattice()

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
		n.plumb.Self.Step(ctx, clk.Tick())

		clock.ApplySpeedNonBlocking(clk, n.plumb.SpeedCh)
		if err := clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}

// openingEmit is everything this node says ONCE, before its loop starts.
//
// Report THIS node's OPENING tilt/normal pair once, before the loop. Self is a
// passive mirror of these (PairNodeSelf.SetTiltIndex) and has no way to derive the
// normal itself, so without this its normal indices sit at their zero value until the
// first arrival or panel click — and since the tilt index opens at 0 too, both
// directions decode to world +y and the two drawn arrows superimpose, which reads as
// the coplanar normal being missing entirely.
func (n *Node) openingEmit() {
	wire.TryEmit(n.plumb.EmitGeometry)
	// This node's own mover-owned startup geometry emit — see Self's own doc comment.
	// There is no separate nodeMover goroutine to make this emit any more.
	n.plumb.Self.EmitGeometryOnce()
	n.syncTiltIndex()
}

// paceOnBeadArrival drains In non-blocking. A bead arrival PACES the exchange and marks the
// round trip; it DECIDES nothing. It used to step this node's tilt one click in this
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
func (n *Node) paceOnBeadArrival() {
	if _, ok := n.plumb.In.PollRecv(); ok {
		if n.plumb.Fire != nil {
			n.plumb.Fire()
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
				plumb: nodePlumbing{Clock: clock.NewRealClock()},
			}
			// This node's own spec id, which is what START is addressed by — see PairID's
			// own doc comment. A name that is not a number leaves PairID at 0, so such a
			// node simply never opens an exchange rather than silently becoming id 1.
			if id, err := strconv.Atoi(a.Name()); err == nil {
				n.plumb.PairID = int32(id)
			}
			n.plumb.Fire = a.Fire()
			if clk := a.Clock(); clk != nil {
				n.plumb.Clock = clk
			}
			n.plumb.SpeedCh = a.SpeedCh()
			n.plumb.In = a.In("In")
			n.plumb.Out = a.Out("Out")
			// The persisted seed is a NUMBER from outside this kind — an old position.json
			// can hold anything, including a running count from before the tilt became a
			// state — so it comes in through seedState, which asks the ring which state
			// carries that index. After this line the tilt is a state and stays one.
			// This node's own lattice, opened at the scene's currently-persisted point
			// count (view/lattice.json via BuildArgs.LatticePointsSeed) rather than the
			// compile-time default.
			latticeSeed := a.LatticePointsSeed()
			n.lattice.Ring = newRing(latticeSeed)
			seed, seedUnknown := n.lattice.Ring.seedState(a.TiltVectorAngleSeed())
			n.setTop(seed)
			n.tilt.TiltEditIn = a.TiltEditIn()
			n.lattice.LatticeIn = a.LatticeIn()
			// Self replaces the old SyncTiltIndex/SyncReceivedVector/ClearOutBeads
			// messages-to-a-separate-mover-goroutine (task/pair-node-owns-itself):
			// this node's own goroutine now owns that mover state directly, so what
			// used to be a message is a plain method call on the same object below.
			self := a.ClaimSelfDrive()
			n.plumb.Self = self
			n.lattice.SyncLatticePoints = func(points int32) {
				self.SetLatticePoints(points)
			}
			n.lattice.SyncLatticePoints(latticeSeed)
			if seedUnknown {
				// The persisted index is not one this ring has — a position.json written
				// before the tilt became a state, or by a build with a different lattice.
				// The node opens at the origin and says which number it refused, rather
				// than computing some other direction and drawing it as if chosen.
				self.Breadcrumb("pair-seed-unknown", fmt.Sprintf(
					"node=%s persisted=%d loaded=%d", a.Name(), a.TiltVectorAngleSeed(), seed.idx))
			}
			n.tilt.SyncTiltIndex = func(theta, normalTheta, bottomTheta int32) {
				self.SetTiltIndex(theta, normalTheta, bottomTheta)
			}
			n.vec.SyncReceivedVector = func(theta int32, set bool) {
				self.SetReceivedVector(theta, set)
			}
			n.plumb.ClearOutBeads = func() { self.ClearOutBeads() }
			n.vec.VectorOut = a.VectorOut()
			n.vec.VectorIn = a.VectorIn()
			// EmitGeometry stays nil deliberately — n.Self.EmitGeometryOnce (Update)
			// makes this node's own startup geometry emit instead.
			return n, nil
		})
}
