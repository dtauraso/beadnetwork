// Package distancegroups holds the "distance home button" toolbar panel's pure math: the
// three named groups of node-pair distances, the current-max/target-length computation, and
// the per-pair reposition loop. Moved out of nodes/Wiring's distance_groups.go (god-object
// decomposition) — the reach-in that used to pin these functions to nodes/Wiring was their
// SIGNATURE (*moverRegistry, *layoutQuantizer, both unexported Wiring actor types), not
// their BODIES: every function here only ever calls two things through those types —
// "read a node's live center" and "move a node's target" — so both are now parameters
// (centerOf/rootMove), bound to the real actor methods at the Wiring call site, the same
// bound-func-value pattern nodes/Wiring already uses elsewhere (move_dispatch_construct.go's
// `ng.msg.sendMove = md.mr.enqueueFor(ng)`).
//
// This is its own package rather than landing in an existing one because its boundary is
// genuinely different from every existing nodes/Wiring subpackage: geom is stateless vector
// math with no domain table, topoderive computes structural facts once at LOAD time from a
// loadspec.TopoSpec, and this is RUNTIME toolbar-panel math (one arrow click) over live
// per-node centers reached through injected accessor/mover functions — no existing package
// is that boundary.
package distancegroups

import (
	"context"
	"time"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// Pair is one (source, target) bead-edge pair; TARGET is the node that moves.
type Pair struct {
	Source string
	Target string
}

// GroupOrder is the WIRE ORDER for the 3 groups (group index 0/1/2 on the editor->Go
// bridge, and the order the 3 Overlay GroupLen* columns are populated in — see
// nodes/Wiring's DistanceGroupLens). Never reorder without updating both.
var GroupOrder = []string{"time", "input", "gate"}

// Groups is the Go-authoritative group->pairs table (CLAUDE.md "Model — read first": Go
// owns the group definitions and the math; TS holds no domain state). These 10 pairs are
// exactly the 10 bead edges in topology/nodes/*/edges/*.json.
var Groups = map[string][]Pair{
	"time":  {{Source: "2", Target: "5"}, {Source: "2", Target: "4"}, {Source: "4", Target: "7"}, {Source: "4", Target: "6"}},
	"input": {{Source: "1", Target: "3"}, {Source: "1", Target: "2"}},
	"gate":  {{Source: "3", Target: "8"}, {Source: "5", Target: "8"}, {Source: "5", Target: "9"}, {Source: "7", Target: "9"}},
}

// Max computes a group's CURRENT max pair length (max over the group's pairs of
// |center(target)-center(source)|), reading live centers through centerOf (bound to
// moverRegistry.centerOfNode by the caller — the same source RootMove/reachRFromPolar use).
// ok is false if the group is unknown, hasGroups is false, or none of its pairs' centers
// are resolvable yet.
//
// hasGroups gates every reader: a scene without these groups resolves NONE of them. This is
// the one gate — both Lens (the VIEW frame's three columns) and ApplyTarget (the arrow-click
// math) come through here, so neither can act on a group belonging to a different scene.
func Max(hasGroups bool, centerOf func(string) (wire.Vec3, bool), group string) (float64, bool) {
	if !hasGroups {
		return 0, false
	}
	pairs, ok := Groups[group]
	if !ok {
		return 0, false
	}
	max := 0.0
	any := false
	for _, p := range pairs {
		cs, okS := centerOf(p.Source)
		ct, okT := centerOf(p.Target)
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

// Lens returns the 3 groups' current max pair lengths, in GroupOrder (time, input, gate) —
// for the VIEW stream's Overlay GroupLenTime/GroupLenInput/GroupLenGate columns. A group
// whose centers aren't resolvable yet reads 0.
func Lens(hasGroups bool, centerOf func(string) (wire.Vec3, bool)) (timeLen, inputLen, gateLen float32) {
	vals := make([]float32, len(GroupOrder))
	for i, g := range GroupOrder {
		if m, ok := Max(hasGroups, centerOf, g); ok {
			vals[i] = float32(m)
		}
	}
	return vals[0], vals[1], vals[2]
}

// ApplyTarget is the controller for one arrow click: groupIdx indexes GroupOrder (0/1/2,
// out of range = no-op); dir > 0 is the up arrow (target length L = currentMax*1.1), dir < 0
// is down (L = currentMax/1.1). For EACH pair in the group's flat list, IN ORDER, the target
// node's new world position is center(source) + normalize(center(target)-center(source))*L,
// applied via rootMove (bound to layoutQuantizer.RootMove by the caller — the same
// decentralized drag entry every programmatic move test uses). Returns false if the group is
// unknown, has no resolvable pair, or groupIdx is out of range.
func ApplyTarget(ctx context.Context, hasGroups bool, centerOf func(string) (wire.Vec3, bool), rootMove func(ctx context.Context, target string, newPos wire.Vec3) bool, groupIdx, dir int) bool {
	if groupIdx < 0 || groupIdx >= len(GroupOrder) {
		return false
	}
	group := GroupOrder[groupIdx]
	pairs, ok := Groups[group]
	if !ok {
		return false
	}
	currentMax, any := Max(hasGroups, centerOf, group)
	if !any {
		return false
	}
	targetLen := currentMax * 1.1
	if dir < 0 {
		targetLen = currentMax / 1.1
	}
	moved := false
	for _, p := range pairs {
		cs, okS := centerOf(p.Source)
		ct, okT := centerOf(p.Target)
		if !okS || !okT {
			continue
		}
		offset := ct.Sub(cs)
		if offset.Length() == 0 {
			continue
		}
		newPos := cs.Add(offset.Normalize().Scale(targetLen))
		if rootMove(ctx, p.Target, newPos) {
			moved = true
			// rootMove is fire-and-forget (see its doc comment at the Wiring call site): it
			// returns before the target's commit lands. A later pair in this SAME group can
			// name this target as its own SOURCE (e.g. "time"'s node 5: target of (2,5),
			// source of (5,8)/(5,7)), so the next iteration's center(source) read must
			// observe the settled position, not a stale pre-move one — settle here, in
			// order, before moving on. Bounded (never blocks forever); a target whose commit
			// doesn't land within the deadline just leaves the next pair reading whatever
			// center is currently live, same as any other cross-goroutine read on this seam.
			waitForCenterSettle(centerOf, p.Target, newPos)
		}
	}
	return moved
}

// waitForCenterSettle polls centerOf(id) until it matches want (within a small tolerance)
// or a short deadline passes. See ApplyTarget's call site.
func waitForCenterSettle(centerOf func(string) (wire.Vec3, bool), id string, want wire.Vec3) {
	const tol = 1e-6
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if c, ok := centerOf(id); ok && c.Sub(want).Length() <= tol {
			return
		}
		time.Sleep(time.Millisecond)
	}
}
