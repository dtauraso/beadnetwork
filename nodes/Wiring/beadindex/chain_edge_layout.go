package beadindex

import (
	"fmt"
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// chain_edge_layout.go — chainBeads' (chain_beads.go) own PER-EDGE geometry, split into the
// phases David's task brief named: resolve target geometry, compute the step count, place
// beads along the chain, decide which are lit, build the breadcrumb text. Each function here
// is a pure function of its arguments — no *nodeGeometry, no channel, no m.* read — lifted
// out of chainBeads' single 265-line loop body once its non-arithmetic statements (the
// partnerCenters/neighborKinds/mutualTargets lookups, the PublishSteps/SendStepsNonBlocking
// sends, the m.tr.Breadcrumb call, the m.reconcileBeadChain call) were set apart from the
// arithmetic sitting between them. chainBeads keeps exactly those non-arithmetic statements
// and calls these three in order, threading their outputs straight through.

// Pulse is one live in-flight traversal on an outgoing edge, gathered by chainBeads from its
// own outgoing wires' LiveBeadFractions(tick) exactly as before: its own continuous [0,1)
// progress T and the STEP COUNT its own T was computed against (Steps travels with the bead
// — see ChainBeadRows' doc comment on why a second, possibly-changed length is never
// substituted for it), plus the bead's own carried value.
type Pulse struct {
	T     float64
	Steps int
	Val   int32
}

// ChainEdgeGeometry is chain_beads.go's own "dist, liveDir, ok := nodegeom.EdgeCenterDistAndDir(...);
// count := nodegeom.EdgeStepCount(...)" pair, unchanged, given a name: the ONE live
// measurement of this edge (docs/bead-model/bead-lattice.md) and the ONE integer bead-step count
// derived from it, read once and reused for layout, the published step count, and the
// breadcrumb's own K. ok is false exactly when EdgeCenterDistAndDir's own coincident-centers
// guard fires — the caller skips this edge exactly as chainBeads' own `if !ok { continue }`
// already does.
func ChainEdgeGeometry(selfCenter, targetCenter wire.Vec3, selfTorusR float64, selfKind, targetKind string) (dist float64, dir wire.Vec3, count int, ok bool) {
	dist, dir, ok = nodegeom.EdgeCenterDistAndDir(selfCenter, targetCenter)
	if !ok {
		return dist, dir, 0, false
	}
	count = nodegeom.EdgeStepCount(dist, selfKind, targetKind)
	return dist, dir, count, true
}

// ChainBeadRows is chainBeads' own two placement loops (the placeholder loop and the
// lit-pulse loop), unchanged arithmetic, given a name. base/step are chainBeads' own
// `selfTorusR + lattice.BeadTorusOuterR` / `lattice.BeadStepR`; dir is the live aim
// (ChainEdgeGeometry's own dir); chainSep is the mutual-pair perpendicular offset
// (nodegeom.ParallelChainOffset, computed by the caller — zero for every ordinary edge).
//
// resolved/resolvedValid are the bead-actor chain's own already-drained snapshot positions
// (edgeBeadChain.last/valid, chain_beads.go/bead_chain.go), parallel to placeholder index i:
// when resolvedValid[i] is true, that bead's own goroutine has already resolved its
// position from the same aim and this function uses it verbatim rather than recomputing the
// identical value; both nil/empty in every bare-literal test nodeMover (beadTickFn nil), in
// which case every index falls back to the formula. This keeps the function pure — it reads
// the actor's ALREADY-DRAINED, already-local values, never the actor itself.
func ChainBeadRows(dir, chainSep wire.Vec3, base, step float64, count int, resolved []wire.Vec3, resolvedValid []bool, pulses []Pulse) (ox, oy, oz []float32, lit []uint8, litVal []int32) {
	// One coordinate: bead index i. Offset from this node's centre is
	// base + i*step (docs/bead-model/bead-lattice.md "Placement"). "Beads never inside a node"
	// falls out of this tangency, with no clamp.
	for i := 0; i < count; i++ {
		var p wire.Vec3
		if i < len(resolvedValid) && resolvedValid[i] && i < len(resolved) {
			// The bead's own goroutine already resolved this position from the broadcast
			// aim (or an earlier one at the same aim) — use it rather than recomputing the
			// identical value here.
			p = resolved[i]
		} else {
			// dir is ALREADY a unit cartesian direction — scaling it by the offset places
			// the bead directly, with no cartesian->polar->cartesian round trip.
			p = dir.Scale(BeadPlacementOffset(base, step, i))
		}
		// Offsets are NODE-LOCAL (the buffer carries them relative to this node's own
		// centre), so the separation is added here in the same local frame rather than
		// being folded into the aim — bending the aim would change the chain's LENGTH and
		// therefore its published step count, making layout and timing disagree.
		p = p.Add(chainSep)
		ox = append(ox, float32(p.X))
		oy = append(oy, float32(p.Y))
		oz = append(oz, float32(p.Z))
		lit = append(lit, 0)
		litVal = append(litVal, 0)
	}
	// THEN one LIT row per live pulse, appended after this edge's chain. Its position is the
	// same lattice arithmetic the placeholders use, evaluated at a CONTINUOUS index instead
	// of an integer i — see PulsePlacementOffset's own doc comment for why pl.Steps (the
	// bead's OWN step count), not this cycle's live count, spans it.
	for _, pl := range pulses {
		p := dir.Scale(PulsePlacementOffset(base, step, pl.T, pl.Steps))
		p = p.Add(chainSep)
		ox = append(ox, float32(p.X))
		oy = append(oy, float32(p.Y))
		oz = append(oz, float32(p.Z))
		lit = append(lit, 1)
		litVal = append(litVal, pl.Val)
	}
	return ox, oy, oz, lit, litVal
}

// ChainAimBreadcrumbText is chainBeads' own "chain-aim" breadcrumb VALUE string, unchanged
// formatting, given a name — pure string formatting from already-local values (to, count,
// dist, dir), separate from the m.tr.Breadcrumb call next to it that actually sends it: that
// call is a side effect (a send on the trace channel), this formatting is not. dir is
// ALREADY a unit vector (ChainEdgeGeometry's own dir), so its own theta/phi come from
// math.Acos/math.Atan2 directly on the unit components — no second vector-length or
// re-normalize call, which chain_beads.go's own file-level sqrt guard
// (tools/network/beads/check-no-sqrt-in-chain-beads.sh) bans in that file (trig itself is
// allowed there too, only the sqrt-fingerprinted helpers are not — this file carries neither
// the guard nor a sqrt: math.Acos/math.Atan2 are not sqrt).
func ChainAimBreadcrumbText(to string, count int, dist float64, dir wire.Vec3) string {
	liveTheta := math.Acos(geom.Clamp(dir.Y, -1, 1))
	livePhi := math.Atan2(dir.Z, dir.X)
	return fmt.Sprintf(
		"to=%s count=%d K=%d liveDir=(theta=%.4f,phi=%.4f)",
		to, count, int(math.Round(dist/lattice.BeadStepR)), liveTheta, livePhi)
}
