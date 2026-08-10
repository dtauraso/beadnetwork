// build_args_tilt_vector.go — BuildArgs methods for a node's own TILT-VECTOR-ANGLE seed and
// edit channel (TiltVectorAngleSeed/TiltEditIn), plus its dedicated tilt-VECTOR channel ends
// (VectorOut/VectorIn) — the (PairNode-only) path where a kind's own goroutine owns its own
// tilt index and vector exchange instead of a separate nodeMover. Split out of
// build_args.go — see that file's header.

package Wiring

import "github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"

// TiltVectorAngleSeed returns this node's persisted tilt-vector-angle index
// (specNode.TopTiltVectorThetaIdx, 0 default) — the load-time seed for a kind that
// owns its OWN index field (PairNode), so it starts from the same persisted value the
// mover used to seed itself with before this reshape.
func (a BuildArgs) TiltVectorAngleSeed() (theta int32) {
	return a.tiltThetaIdx
}

// TiltEditIn claims this node's dedicated inbound channel for a panel-driven tilt-angle
// click (TiltVectorAnglePanel), registering it in MoveDispatch.tiltEditIns so
// applyUpdateTiltVector (stdin_reader.go) routes that node's edits HERE instead of to its
// mover. Call this ONLY from a kind whose own goroutine independently owns/decides its
// tilt index (PairNode) — every other kind must keep using the old mover-owned path by
// simply never calling this. nil-safe: a.pb.md is nil on a bare test build with no
// loader, in which case this returns a channel that is never written to (PollRecv-style
// non-blocking reads on it always find nothing, matching every other build-time fallback
// in this file).
func (a BuildArgs) TiltEditIn() <-chan TiltEditMsg {
	md := a.pb.md
	if md == nil {
		return make(chan TiltEditMsg)
	}
	panelToNodeTiltEditIn := make(chan TiltEditMsg, moverInboxDepth)
	if md.inboxes.tiltEdit == nil {
		md.inboxes.tiltEdit = map[string]chan TiltEditMsg{}
	}
	md.inboxes.tiltEdit[a.name] = panelToNodeTiltEditIn
	return panelToNodeTiltEditIn
}

// VectorOut returns this node's own SEND end of its dedicated tilt-vector channel
// (tilt_vector_channel.go) — the buffered-1, latest-wins, non-blocking channel
// carrying a TiltVectorMsg alongside the ordinary bead edge, wired only when this
// node's outgoing edge's OTHER endpoint also asked for one (build.go's
// allocateVectorChannels). nil when unwired (every kind but PairNode, a node
// with no outgoing vector-capable edge, or a bare test build with no loader) —
// SendVectorLatestNonBlocking already treats a nil channel as a no-op send, the
// same fallback shape as every other unwired-port case in this file.
func (a BuildArgs) VectorOut() chan<- tiltvector.TiltVectorMsg {
	if a.pb.vectorOut == nil {
		return nil
	}
	return a.pb.vectorOut[a.name]
}

// VectorIn returns this node's own RECEIVE end of its dedicated tilt-vector channel
// — the counterpart to VectorOut. nil when unwired, in which case
// PollRecvVector(nil) always reports ok=false, matching every other unwired-port
// fallback in this file.
func (a BuildArgs) VectorIn() <-chan tiltvector.TiltVectorMsg {
	if a.pb.vectorIn == nil {
		return nil
	}
	return a.pb.vectorIn[a.name]
}
