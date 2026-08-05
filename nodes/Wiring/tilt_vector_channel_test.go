package Wiring

import "testing"

// SendVectorLatestNonBlocking must never block: a full buffer is a drop, not a wait —
// this is a single call's own behavior, no second goroutine involved.
func TestSendVectorLatestNonBlockingDropsWhenFull(t *testing.T) {
	ch := make(chan TiltVectorMsg, 1)
	SendVectorLatestNonBlocking(ch, TiltVectorMsg{ThetaIdx: 1, PhiIdx: 1})
	// Buffer now full; this second send must return immediately rather than block,
	// and — since nothing drains between the two sends — must NOT overwrite the
	// first pending value (this call has no way to drain a send-only channel end;
	// only the receiver may drain).
	SendVectorLatestNonBlocking(ch, TiltVectorMsg{ThetaIdx: 2, PhiIdx: 2})
	got := <-ch
	if got.ThetaIdx != 1 || got.PhiIdx != 1 {
		t.Fatalf("want the first pending value preserved (1,1), got (%d,%d)", got.ThetaIdx, got.PhiIdx)
	}
}

// A nil channel is "nothing wired" — the send must be a no-op, not a panic.
func TestSendVectorLatestNonBlockingNilIsNoop(t *testing.T) {
	SendVectorLatestNonBlocking(nil, TiltVectorMsg{ThetaIdx: 1})
}

// PollRecvVector on an empty channel returns immediately with ok=false.
func TestPollRecvVectorEmptyReturnsFalse(t *testing.T) {
	ch := make(chan TiltVectorMsg, 1)
	if _, ok := PollRecvVector(ch); ok {
		t.Fatalf("want ok=false on an empty channel")
	}
}

// PollRecvVector on a nil channel ("nothing wired") also returns ok=false, not a block.
func TestPollRecvVectorNilReturnsFalse(t *testing.T) {
	if _, ok := PollRecvVector(nil); ok {
		t.Fatalf("want ok=false on a nil channel")
	}
}

// PollRecvVector drains a pending value.
func TestPollRecvVectorDrainsPending(t *testing.T) {
	ch := make(chan TiltVectorMsg, 1)
	ch <- TiltVectorMsg{ThetaIdx: 5, PhiIdx: -2}
	v, ok := PollRecvVector(ch)
	if !ok || v.ThetaIdx != 5 || v.PhiIdx != -2 {
		t.Fatalf("want ok=true value (5,-2), got ok=%v value (%d,%d)", ok, v.ThetaIdx, v.PhiIdx)
	}
}

// KindWantsVectorChannel only names Node1/Node2 today — every other kind gets nothing.
func TestKindWantsVectorChannelOnlyNode1Node2(t *testing.T) {
	if !KindWantsVectorChannel("Node1") || !KindWantsVectorChannel("Node2") {
		t.Fatalf("want Node1 and Node2 to ask for a vector channel")
	}
	if KindWantsVectorChannel("Pulse") || KindWantsVectorChannel("Time") || KindWantsVectorChannel("") {
		t.Fatalf("want no other kind to ask for a vector channel")
	}
}
