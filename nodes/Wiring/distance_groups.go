// distance_groups.go — the "distance home button" toolbar panel's Go-owned controller.
//
// Three named groups of node-pair distances (GATE definitions below, node ids are the
// string ids from topology/nodes/<id>). Each pair is (source, target) taken from the
// bead edge's direction; the TARGET node is the one that moves. Clicking a group's up
// arrow sets that group's target length L = currentMax*1.1 (down: currentMax/1.1),
// where currentMax is the max over the group's pairs of |center(target)-center(source)|
// (mirrors reachRFromPolar's max-over-edges loop, quantized_move.go). Then EVERY pair in
// the group is repositioned so its length becomes L, in FLAT LIST ORDER — a target node
// that appears in two pairs (gate group: 9 is the target of both (3,9) and (6,9); 10 of
// both (6,10) and (8,10)) ends at the LAST pair's placement. This is intended and
// accepted (per the agreed model): there is no tree/graph solver, no averaging, no
// equal-radii resolve for a shared target.
//
// The repositioning itself reuses md.RootMove exactly as the drag tests call it
// programmatically (abc_drag_count_target_node_test.go) — RootMove already routes the
// move to the target node's OWN goroutine, commits via commitNodeMoveLocal, and
// rebroadcasts geometry so incident edges' segments recompute and redraw. This file adds
// no new position/commit path.
package Wiring

import "time"

// distancePair is one (source, target) bead-edge pair; TARGET is the node that moves.
type distancePair struct {
	Source string
	Target string
}

// distanceGroupOrder is the WIRE ORDER for the 3 groups (group index 0/1/2 on the
// editor->Go bridge, and the order the 3 Overlay GroupLen* columns are populated in —
// see MoveDispatch.DistanceGroupLens). Never reorder without updating both.
var distanceGroupOrder = []string{"time", "input", "gate"}

// distanceGroups is the Go-authoritative group->pairs table (CLAUDE.md "Model — read
// first": Go owns the group definitions and the math; TS holds no domain state). These
// 10 pairs are exactly the 10 bead edges in topology/edges/*.json.
var distanceGroups = map[string][]distancePair{
	"time":  {{Source: "2", Target: "6"}, {Source: "2", Target: "5"}, {Source: "5", Target: "8"}, {Source: "5", Target: "7"}},
	"input": {{Source: "1", Target: "3"}, {Source: "1", Target: "2"}},
	"gate":  {{Source: "3", Target: "9"}, {Source: "6", Target: "9"}, {Source: "6", Target: "10"}, {Source: "8", Target: "10"}},
}

// distanceGroupMax computes a group's CURRENT max pair length (max over the group's
// pairs of |center(target)-center(source)|), reading live centers from md's own
// centerMirror (md.centerOfNode — the same source RootMove/reachRFromPolar use). ok is
// false if the group is unknown or none of its pairs' centers are resolvable yet.
func (md *MoveDispatch) distanceGroupMax(group string) (float64, bool) {
	pairs, ok := distanceGroups[group]
	if !ok {
		return 0, false
	}
	max := 0.0
	any := false
	for _, p := range pairs {
		cs, okS := md.centerOfNode(p.Source)
		ct, okT := md.centerOfNode(p.Target)
		if !okS || !okT {
			continue
		}
		if d := ct.Sub(cs).Length(); d > max {
			max = d
		}
		any = true
	}
	return max, any
}

// DistanceGroupLens returns the 3 groups' current max pair lengths, in
// distanceGroupOrder (time, input, gate) — for the VIEW stream's Overlay
// GroupLenTime/GroupLenInput/GroupLenGate columns (read-only reflect; see
// view_stream.go's emitViewFrame). A group whose centers aren't resolvable yet reads 0.
func (md *MoveDispatch) DistanceGroupLens() (timeLen, inputLen, gateLen float32) {
	vals := make([]float32, len(distanceGroupOrder))
	for i, g := range distanceGroupOrder {
		if m, ok := md.distanceGroupMax(g); ok {
			vals[i] = float32(m)
		}
	}
	return vals[0], vals[1], vals[2]
}

// ApplyDistanceGroupTarget is the controller for one arrow click: groupIdx indexes
// distanceGroupOrder (0/1/2, out of range = no-op); dir > 0 is the up arrow (target
// length L = currentMax*1.1), dir < 0 is down (L = currentMax/1.1). For EACH pair in
// the group's flat list, IN ORDER, the target node's new world position is
// center(source) + normalize(center(target)-center(source))*L, applied via
// md.RootMove(target, newPos) — the same decentralized drag entry every programmatic
// move test uses. Returns false if the group is unknown, has no resolvable pair, or
// groupIdx is out of range.
func (md *MoveDispatch) ApplyDistanceGroupTarget(groupIdx, dir int) bool {
	if groupIdx < 0 || groupIdx >= len(distanceGroupOrder) {
		return false
	}
	group := distanceGroupOrder[groupIdx]
	pairs, ok := distanceGroups[group]
	if !ok {
		return false
	}
	currentMax, any := md.distanceGroupMax(group)
	if !any {
		return false
	}
	targetLen := currentMax * 1.1
	if dir < 0 {
		targetLen = currentMax / 1.1
	}
	moved := false
	for _, p := range pairs {
		cs, okS := md.centerOfNode(p.Source)
		ct, okT := md.centerOfNode(p.Target)
		if !okS || !okT {
			continue
		}
		offset := ct.Sub(cs)
		if offset.Length() == 0 {
			continue
		}
		newPos := cs.Add(offset.Normalize().Scale(targetLen))
		if md.RootMove(p.Target, newPos) {
			moved = true
			// RootMove is fire-and-forget (a moveMsgKindDrag message to the target's OWN
			// goroutine — see its doc comment): it returns before the target's commit
			// lands. A later pair in this SAME group can name this target as its own
			// SOURCE (e.g. "time"'s node 5: target of (2,5), source of (5,8)/(5,7)), so
			// the next iteration's center(source) read must observe the settled position,
			// not a stale pre-move one — settle here, in order, before moving on. Bounded
			// (never blocks forever); a target whose commit doesn't land within the
			// deadline just leaves the next pair reading whatever center is currently
			// live, same as any other cross-goroutine read on this seam.
			waitForCenterSettle(md, p.Target, newPos)
		}
	}
	return moved
}

// waitForCenterSettle polls md.centerOfNode(id) until it matches want (within a small
// tolerance) or a short deadline passes. See ApplyDistanceGroupTarget's call site.
func waitForCenterSettle(md *MoveDispatch, id string, want vec3) {
	const tol = 1e-6
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if c, ok := md.centerOfNode(id); ok && c.Sub(want).Length() <= tol {
			return
		}
		time.Sleep(time.Millisecond)
	}
}
