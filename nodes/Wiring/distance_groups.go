// distance_groups.go — the "distance home button" toolbar panel's Go-owned controller.
//
// Three named groups of node-pair distances (GATE definitions below, node ids are the
// string ids from topology/nodes/<id>). Each pair is (source, target) taken from the
// bead edge's direction; the TARGET node is the one that moves. Clicking a group's up
// arrow sets that group's target length L = currentMax*1.1 (down: currentMax/1.1),
// where currentMax is the max over the group's pairs of |center(target)-center(source)|
// (mirrors reachRFromPolar's max-over-edges loop, quantized_move.go). Then EVERY pair in
// the group is repositioned so its length becomes L, in FLAT LIST ORDER — a target node
// that appears in two pairs (gate group: 8 is the target of both (3,8) and (5,8); 9 of
// both (5,9) and (7,9)) ends at the LAST pair's placement. This is intended and
// accepted (per the agreed model): there is no tree/graph solver, no averaging, no
// equal-radii resolve for a shared target.
//
// This file computes NO positions. It hands each edge a single number — the target length
// — and the edge, which is the only thing that knows both its endpoints, works out the
// displacement and sends it to the endpoint that moves. That endpoint applies the delta to
// itself and commits through the same owner-goroutine commit path a drag uses
// (commitNodeMoveLocal), so this file still adds no new position/commit path; it simply
// stopped being the one doing the arithmetic.
package Wiring

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
// 10 pairs are exactly the 10 bead edges in topology/nodes/*/edges/*.json.
var distanceGroups = map[string][]distancePair{
	"time":  {{Source: "2", Target: "5"}, {Source: "2", Target: "4"}, {Source: "4", Target: "7"}, {Source: "4", Target: "6"}},
	"input": {{Source: "1", Target: "3"}, {Source: "1", Target: "2"}},
	"gate":  {{Source: "3", Target: "8"}, {Source: "5", Target: "8"}, {Source: "5", Target: "9"}, {Source: "7", Target: "9"}},
}

// distanceGroupMax returns a group's CURRENT max pair length, as a REDUCTION over lengths
// each edge measured and published itself (moverRegistry.lengthOfPair, fed by every
// edgeMover's own publishLength). ok is false if the group is unknown or none of its pairs
// has a published length yet.
//
// It used to derive each length here, reading both endpoints out of centerMirror and
// subtracting. That was the wrong owner twice over: dispatch does not own either node's
// position, and centerMirror is documented as EVENTUALLY CONSISTENT and "acceptable for
// camera/framing reads, which is the only remaining caller class" — while this caller was
// not a framing read but the input to a position computation. The gap between those two
// facts is exactly what waitForCenterSettle was added to paper over.
//
// Reducing over owner-published values does not have that problem. A late value is an
// older LENGTH — a distance those two endpoints really did have — never a distance
// between two positions that never coexisted, which is what subtracting two independently
// stale centers can produce.
func (md *MoveDispatch) distanceGroupMax(group string) (float64, bool) {
	pairs, ok := distanceGroups[group]
	if !ok {
		return 0, false
	}
	max := 0.0
	any := false
	for _, p := range pairs {
		d, okL := md.mr.lengthOfPair(p.Source, p.Target)
		if !okL {
			continue
		}
		if d > max {
			max = d
		}
		any = true
	}
	return max, any
}

// DistanceGroupLens was REMOVED. The 3 Overlay GroupLenTime/GroupLenInput/GroupLenGate
// columns it fed are gone with it: they lived on the VIEW frame, which is EVENT-DRIVEN
// (view_stream.go), so any value computed there refreshes only when something unrelated
// happens to emit — untenable once the moves themselves are asynchronous. Each edge now
// carries its own Len and GroupIdx on its own per-owner stream frame, which flows when
// that edge moves, and the panel reduces per group from those.

// distanceGroupIdxForPair returns the index into distanceGroupOrder of the group holding
// the (src,dst) pair, or -1 when the pair is in no group. It is the ONE place the
// Go-authoritative group table is projected onto an edge, so each edgeMover can stamp its
// own membership on its own stream frame (Buffer bufLayoutEdge.GroupIdx) without TS ever
// holding a membership table.
//
// Static: distanceGroups is a package-level literal, never mutated, so this is a pure
// lookup safe to call from any mover's own goroutine.
func distanceGroupIdxForPair(src, dst string) int32 {
	for i, g := range distanceGroupOrder {
		for _, p := range distanceGroups[g] {
			if p.Source == src && p.Target == dst {
				return int32(i)
			}
		}
	}
	return -1
}

// ApplyDistanceGroupTarget is the controller for one arrow click: groupIdx indexes
// distanceGroupOrder (0/1/2, out of range = no-op); dir > 0 is the up arrow (target length
// L = currentMax*1.1), dir < 0 is down (L = currentMax/1.1). It then hands EVERY edge in
// the group that one number and returns. Returns false if the group is unknown, has no
// resolvable length, groupIdx is out of range, or no edge accepted the message.
//
// It computes no positions. Previously it walked the pairs IN ORDER, and for each one read
// center(source) and center(target) off dispatch's mirror, computed
// center(source) + normalize(center(target)-center(source))*L, and handed that absolute
// position to RootMove. Three consequences, all of which are gone:
//
//   - It had to read two foreign nodes' positions to do the arithmetic, from a mirror
//     documented as eventually consistent and meant for framing reads.
//   - Because a later pair's SOURCE could be an earlier pair's TARGET (node 5 is the
//     target of (2,5) and the source of (5,8)/(5,7)), the loop had to observe each move
//     land before computing the next — waitForCenterSettle, a 200ms poll that returned
//     silently on timeout and left the next pair reading a stale center.
//   - It emitted a VIEW frame at the end, because the panel's lengths were computed inside
//     that emit. Those columns now ride each edge's own stream (Buffer bufLayoutEdge.Len),
//     so there is nothing here to refresh.
//
// Now: each edge is told its target length and works out its own endpoint displacement
// from geometry it already owns; the endpoint applies that delta to itself. Pair order
// stops mattering, because nothing here reads a position that another pair might change.
//
// A node shared by two edges in the group (gate: 8 is the target of both (3,8) and (5,8))
// receives two deltas and applies both, in whatever order they arrive at its own inbox,
// ending where the second one puts it. That is the same "last pair wins, no solver, no
// averaging" outcome the ordered loop had, now reached by the owner arbitrating its own
// mail rather than by dispatch sequencing on its behalf.
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
	sent := false
	for _, p := range pairs {
		if md.mr.sendEdgeSetLength(p.Source, p.Target, targetLen) {
			sent = true
		}
	}
	return sent
}
