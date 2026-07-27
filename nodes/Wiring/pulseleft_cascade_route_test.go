// pulseleft_cascade_route_test.go — PulseLeft's cascade delta rules.
//
// PulseLeft (node 3) is a cascade TERMINUS with a sender whitelist:
//   - it ATTENDS to a delta triple only from an Input or SelectRight
//     ("WindowAndInhibitLeftGate") cascade neighbor — any other sender kind is dropped
//     outright in the moveMsgKindDeltaForward handler (no record, no relay);
//   - it NEVER relays a triple onward, from any sender and even as a direct drag
//     recipient (the guard at the top of forwardDelta).

package Wiring

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// TestForwardDeltaPulseLeftNeverRelays: whatever the sender kind, a PulseLeft forwards to
// nobody — it is the end of the cascade.
func TestForwardDeltaPulseLeftNeverRelays(t *testing.T) {
	for _, sender := range []string{"1", "8"} {
		var mu sync.Mutex
		var got []string

		nm := &nodeMover{
			id:           "3",
			selfKind:     "PulseLeft",
			cascadeEdges: []string{"1", "8"},
			cascadeKinds: map[string]string{"1": "Input", "8": "WindowAndInhibitLeftGate"},
			sendMove: func(id string, msg moveMsg) {
				mu.Lock()
				got = append(got, id)
				mu.Unlock()
			},
		}

		nm.forwardDelta(nil, sender, 1, 2, 3)

		mu.Lock()
		if len(got) != 0 {
			t.Errorf("PulseLeft relaying a delta from %q: want no forwards (terminus), got %v", sender, got)
		}
		mu.Unlock()
	}
}

// TestPulseLeftAttendsOnlyInputAndSelectRight exercises the handler-level whitelist: a
// forward from Input or WindowAndInhibitLeftGate is recorded (gotForwardMsg==1); one from
// any other kind leaves this node's forward state untouched.
func TestPulseLeftAttendsOnlyInputAndSelectRight(t *testing.T) {
	cases := []struct {
		senderID   string
		senderKind string
		want       uint8
	}{
		{"1", "Input", 1},
		{"8", "WindowAndInhibitLeftGate", 1},
		{"4", "Time", 0},
		{"5", "Pulse", 0},
	}
	for _, c := range cases {
		nm := &nodeMover{
			id:           "3",
			selfKind:     "PulseLeft",
			cascadeEdges: []string{"1", "8"},
			cascadeKinds: map[string]string{"1": "Input", "8": "WindowAndInhibitLeftGate", "4": "Time", "5": "Pulse"},
			sendMove:     func(id string, msg moveMsg) {},
		}
		nm.handle(moveMsg{Kind: moveMsgKindDeltaForward, NodeID: "3", SenderID: c.senderID,
			DeltaA: 1, DeltaB: 2, DeltaC: 3})

		if nm.gotForwardMsg != c.want {
			t.Errorf("PulseLeft delta from %s (%s): gotForwardMsg=%d, want %d",
				c.senderID, c.senderKind, nm.gotForwardMsg, c.want)
		}
	}
}

// TestPulseLeftDragTerminatesCascade drags node 5 (Pulse) on the REAL production topology.
// The delta reaches node 3 (PulseLeft) from node 8 (WindowAndInhibitLeftGate = SelectRight)
// along 5->8->3, which node 3 attends to — but node 3 must not forward it on to node 1
// (Input), which is what used to close the cascade cycle 5->8->3->1->2.
func TestPulseLeftDragTerminatesCascade(t *testing.T) {
	root := filepath.Join(repoRootForDeltaForwardTest(t), "topology")
	tr := T.NewWithSinkHook(nil, nil)

	_, _, md, _, err := LoadTopology(context.Background(), root, tr, wire.NewRealClock())
	if err != nil {
		t.Fatalf("LoadTopology: %v", err)
	}

	for _, id := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"} {
		_ = wireNodeStream(t, md, id)
		if nm, ok := md.mr.nodeMovers[id]; ok {
			nm.nodeRowFor = md.NodeRowFor
			ownMover := nm
			nm.forwardOnce = func(exceptID string, dA, dB, dC int32) {
				ownMover.forwardDelta(md, exceptID, dA, dB, dC)
			}
		}
	}

	var mu sync.Mutex
	fromThree := map[string]bool{}
	md.SetMsgTap(func(destID string, msg moveMsg) {
		if msg.Kind != moveMsgKindDeltaForward || msg.SenderID != "3" {
			return
		}
		mu.Lock()
		fromThree[destID] = true
		mu.Unlock()
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	md.Start(ctx)

	before, ok := md.centerOfNode("5")
	if !ok {
		t.Fatal("no center for 5")
	}
	target := before.Add(vec3{X: 45, Y: -30, Z: 20})
	md.resetAbcDrag()
	if !md.RootMove("5", target) {
		t.Fatal("RootMove(5) returned false")
	}
	pollDragConverged(t, md, "5", target)
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	got := []string{}
	for to := range fromThree {
		got = append(got, to)
	}
	mu.Unlock()

	if len(got) != 0 {
		t.Errorf("node 3 (PulseLeft) forwarded to %v; want none — PulseLeft is a cascade terminus", got)
	}
}
