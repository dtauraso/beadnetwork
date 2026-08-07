// tilt_vector_channel.go — a dedicated node-to-node channel carrying an INTEGER
// angle-index pair (θ, φ in CurveParamTiltVectorAngleStep steps — never floats on a
// channel, per memory/feedback_abc_times_constant_not_rederive.md), alongside the
// existing bead edge. It is used only by the kind that asks for it (today:
// Node1, see vectorCapableKinds below); every other kind gets nothing and is
// entirely unaffected.
//
// Buffered depth 1, latest-wins, both ends non-blocking — same shape as the
// speed-delivery channel (wire.SendSpeedNonBlocking/ApplySpeedNonBlocking) and the
// held-value channel (wire.SendLatestNonBlocking): a send that cannot proceed drops
// rather than blocks, a receive that finds nothing returns immediately.
package Wiring

// TiltVectorMsg is the vector-channel payload: an integer angle index. θ is an
// index × CurveParamTiltVectorAngleStep (π/12) — the boundary conversion to a float
// angle happens only where geometry is rendered/persisted, never on this channel.
// There is no φ: every tilt vector in this exchange lives in the θ-only plane, so a
// second angle would be a free axis with nothing to vary it (see TiltVectorIsAcute).
type TiltVectorMsg struct {
	ThetaIdx int32
	// Points is the LATTICE THE INDEX IS ON — how many directions the sender's ring had when
	// it picked ThetaIdx. An index means nothing without it: 6 is a quarter turn on a
	// 24-point lattice and a half turn on a 12-point one.
	//
	// It exists because the point count is a scene setting a user can change, and the two
	// ends of a pair do not adopt a new one at the same instant — each is its own goroutine,
	// reached by its own message. In the window between, a direction picked on the old
	// lattice can arrive at a node already running the new one. Carrying the count makes
	// that arrival RECOGNISABLE: the receiver drops what is not from its own lattice, which
	// is a definite answer rather than either acting on a number that means something else
	// or dying on it (nodes/Node1's arrivedState).
	//
	// Zero means "unstated", which a bare test build sends; a receiver treats that as its
	// own lattice rather than rejecting it, so a test constructing a message by hand does
	// not have to know this field exists.
	Points int32
	// Reset marks this as a RESET rather than a direction to act on: the receiver zeroes its
	// own tilt and does not reply, which ends the exchange instead of restarting it.
	//
	// It does NOT evict anything already queued — a send-only end cannot drain itself (in Go
	// only a receiver empties a channel), so SendVectorLatestNonBlocking DROPS when the
	// buffer is full rather than replacing. Emptying both directions is done by draining,
	// and it works because the reset reaches EVERY node with a vector: each node drains the
	// one receive end it owns, and between them that is both channels.
	Reset bool
	// Machine names WHICH STATE MACHINE the pair runs, decided from THE TILT BEING SET at the
	// moment of a ▲/▼ click and carried to the other end so both run the same one. It is not a
	// direction to act on: a receiver takes up the machine and steps nothing.
	//
	// TiltMachineNone on every message that is not carrying a choice — an ordinary vector
	// exchange message says nothing about which machine anyone is running.
	Machine TiltMachine
}

// TiltMachine names which state machine a pair node runs. The pair kind maps these to its own
// two machines (nodes/Node1/perpendicular.go, parallel.go); this package names them without
// knowing how either one steps.
//
// WHICH ONE IS DECIDED FROM THE TILT BEING SET, at the click that sets it — nothing is
// remembered to make that decision, no seed and no tally of clicks. The RESET button removes
// the choice; the next click makes a new one.
type TiltMachine int8

const (
	// TiltMachineNone: run neither — what a reset restores, and what an ordinary vector
	// message carries.
	TiltMachineNone TiltMachine = iota
	// TiltMachinePerpendicular: the two tilts a quarter turn apart.
	TiltMachinePerpendicular
	// TiltMachineParallel: the two tilts on the SAME LINE — pointing the same way, or exactly
	// opposite. The rule does not distinguish the two, and both come up (nodes/Node1/parallel.go).
	TiltMachineParallel
)

func (m TiltMachine) String() string {
	switch m {
	case TiltMachinePerpendicular:
		return "perpendicular"
	case TiltMachineParallel:
		return "parallel"
	}
	return "none"
}

// vectorCapableKinds names every kind that asks for a vector channel on its edges.
// A kind not listed here gets no channel at all — allocateVectorChannels below skips
// any edge unless BOTH endpoints are listed.
var vectorCapableKinds = map[string]bool{
	"Node1": true,
}

// KindWantsVectorChannel reports whether kind participates in the vector-channel
// exchange. Exported so build.go's phase helper (in the same package) reads as a
// named predicate rather than an inline map probe.
func KindWantsVectorChannel(kind string) bool {
	return vectorCapableKinds[kind]
}

// SendVectorLatestNonBlocking is the TiltVectorMsg twin of wire.SendLatestNonBlocking
// / wire.SendSpeedNonBlocking: non-blocking drain-then-send onto a buffered-1,
// latest-wins channel. ch must be a channel only this call's caller ever sends on —
// each directed edge's vector channel has exactly one sender, the source node's own
// goroutine, so that invariant holds by construction.
func SendVectorLatestNonBlocking(ch chan<- TiltVectorMsg, v TiltVectorMsg) {
	if ch == nil {
		return
	}
	select {
	case ch <- v:
		return
	default:
	}
	// Buffer full: this is a send-only channel end here, so there is no way for this
	// goroutine to drain its own stale value — only the receiver drains. A full
	// buffer means the receiver hasn't caught up yet; dropping this send (rather
	// than blocking) is correct latest-wins behavior for the RECEIVER's next poll,
	// which will still see the older-but-undrained value now, and the next
	// non-blocking send after that will land once the receiver has caught up.
}

// PollRecvVector is the TiltVectorMsg twin of wire.In.PollRecv: a non-blocking
// receive that returns immediately with ok=false when nothing is pending. ch may be
// nil (an edge whose endpoints did not both ask for a vector channel), in which case
// this always reports ok=false, matching every other unwired-port fallback in this
// package.
func PollRecvVector(ch <-chan TiltVectorMsg) (TiltVectorMsg, bool) {
	if ch == nil {
		return TiltVectorMsg{}, false
	}
	select {
	case v := <-ch:
		return v, true
	default:
		return TiltVectorMsg{}, false
	}
}

// HalfTurnThetaIdx is a half turn (180°) counted in CurveParamTiltVectorAngleStep steps —
// twice PerpendicularThetaIdx's quarter turn. Adding it to a direction's θ index REVERSES
// that direction exactly: the drawn direction is (sinθ, cosθ, 0) (TiltVectors.tsx's
// writeArrowInto with φ fixed at 0), and sin(θ+π) = −sinθ with cos(θ+π) = −cosθ negates
// both components at once.
const HalfTurnThetaIdx = 2 * PerpendicularThetaIdx

// FullTurnThetaIdx is a full turn (360°) counted in CurveParamTiltVectorAngleStep steps —
// twice HalfTurnThetaIdx.
const FullTurnThetaIdx = 2 * HalfTurnThetaIdx

// THE ANGLE TESTS ARE NOT HERE. How far apart two directions are, and which way to step to
// close on a resting state, are decided by the pair kind itself on its own ring of states
// (nodes/Node1/ring.go's separation/missBy/stepToward). This package held an arithmetic
// version, subtracting two indices and reducing the difference onto the lattice; it has no
// callers now, and a second definition of the same question is exactly the thing that drifts.
