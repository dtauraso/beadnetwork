package Wiring

import (
	"context"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// TestTimeStartDragPulseForwardsToTimeOnly drags node 5 (Pulse) on the REAL production
// topology and asserts node 2 (TimeStart) forwards the delta ONLY to node 4 (Time). The
// delta reaches node 2 twice — directly from Pulse(5) and, via the cascade cycle
// 5->8->3->1->2, from Input(1). The Pulse arrival forwards to Time only (forwardDelta's
// rule); the Input arrival is IGNORED (the moveMsgKindDeltaForward handler's TimeStart <-
// Input stop) — so node 2 never leaks the triple to 1 (Input) or back to 5 (Pulse).
func TestTimeStartDragPulseForwardsToTimeOnly(t *testing.T) {
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
	fromTwo := map[string]bool{}
	md.SetMsgTap(func(destID string, msg moveMsg) {
		if msg.Kind != moveMsgKindDeltaForward || msg.SenderID != "2" {
			return
		}
		mu.Lock()
		fromTwo[destID] = true
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
	for to := range fromTwo {
		got = append(got, to)
	}
	mu.Unlock()
	sort.Strings(got)

	if fromTwo["1"] || fromTwo["5"] {
		t.Errorf("node 2 (TimeStart) forwarded to %v; want only [4] — must not leak to 1 (Input) or 5 (Pulse)", got)
	}
	if !fromTwo["4"] {
		t.Errorf("node 2 (TimeStart) did not forward to 4 (Time); got %v", got)
	}
}

// TestTimeStartDragTimeForwardsToPulseOnly drags node 4 (Time) and asserts node 2
// (TimeStart) forwards the delta ONLY to node 5 (Pulse) — the from-Time -> Pulse routing
// rule — never to 1 (Input) or back to 4 (Time).
func TestTimeStartDragTimeForwardsToPulseOnly(t *testing.T) {
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
	fromTwo := map[string]bool{}
	md.SetMsgTap(func(destID string, msg moveMsg) {
		if msg.Kind != moveMsgKindDeltaForward || msg.SenderID != "2" {
			return
		}
		mu.Lock()
		fromTwo[destID] = true
		mu.Unlock()
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	md.Start(ctx)

	before, ok := md.centerOfNode("4")
	if !ok {
		t.Fatal("no center for 4")
	}
	target := before.Add(vec3{X: 45, Y: -30, Z: 20})
	md.resetAbcDrag()
	if !md.RootMove("4", target) {
		t.Fatal("RootMove(4) returned false")
	}
	pollDragConverged(t, md, "4", target)
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	got := []string{}
	for to := range fromTwo {
		got = append(got, to)
	}
	mu.Unlock()
	sort.Strings(got)

	if fromTwo["1"] || fromTwo["4"] {
		t.Errorf("node 2 (TimeStart) forwarded to %v; want only [5] — must not leak to 1 (Input) or 4 (Time)", got)
	}
	if !fromTwo["5"] {
		t.Errorf("node 2 (TimeStart) did not forward to 5 (Pulse); got %v", got)
	}
}
