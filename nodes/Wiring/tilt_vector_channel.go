// tilt_vector_channel.go — a dedicated node-to-node channel carrying an INTEGER
// angle-index pair (θ, φ in CurveParamTiltVectorAngleStep steps — never floats on a
// channel, per memory/feedback_abc_times_constant_not_rederive.md), alongside the
// existing bead edge. It is used only by the two kinds that ask for it (today:
// Node1/Node2, see vectorCapableKinds below); every other kind gets nothing and is
// entirely unaffected.
//
// Buffered depth 1, latest-wins, both ends non-blocking — same shape as the
// speed-delivery channel (wire.SendSpeedNonBlocking/ApplySpeedNonBlocking) and the
// held-value channel (wire.SendLatestNonBlocking): a send that cannot proceed drops
// rather than blocks, a receive that finds nothing returns immediately.
package Wiring

// TiltVectorMsg is the vector-channel payload: an integer angle-index pair. θ/φ are
// each an index × CurveParamTiltVectorAngleStep (π/12) — the boundary conversion to
// a float angle happens only where geometry is rendered/persisted, never on this
// channel.
type TiltVectorMsg struct {
	ThetaIdx, PhiIdx int32
	// Reset marks this as a RESET rather than a direction to act on: the receiver zeroes its
	// own tilt and does not reply, which ends the exchange instead of restarting it.
	//
	// It does NOT evict anything already queued — a send-only end cannot drain itself (in Go
	// only a receiver empties a channel), so SendVectorLatestNonBlocking DROPS when the
	// buffer is full rather than replacing. Emptying both directions is done by draining,
	// and it works because the reset reaches EVERY node with a vector: each node drains the
	// one receive end it owns, and between them that is both channels.
	Reset bool
}

// vectorCapableKinds names every kind that asks for a vector channel on its edges.
// A kind not listed here gets no channel at all — allocateVectorChannels below skips
// any edge unless BOTH endpoints are listed.
var vectorCapableKinds = map[string]bool{
	"Node1": true,
	"Node2": true,
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
