package Wiring

import (
	"math"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// chain_beads.go — the node-owned placeholder bead chain that IS the visual of an edge.
// Design and staging: docs/beads-are-the-edge.md. The LENGTH model (one integer bead-step
// count, tangent placement, no arc) is docs/bead-lattice.md — this file implements it.
//
// A node owns one chain per OUTGOING edge, following ownership the tree already has: an edge
// is stored at topology/nodes/<source>/edges/<label>.json, outgoing only, "carries no
// `source` key: that is the directory it sits in" (.claude/rules/persistence-ownership.md).
// So the source node owns the whole chain — there is no split at a midpoint, which would
// force the lit index to hand off between two nodes' sequences mid-traversal.
//
// TWO THINGS THAT ARE NOT THE SAME, and conflating them is the mistake to avoid:
//
//   - the CHANNELS between nodes are the real connection, carrying delivery and the
//     chain-maintenance messages. They are never drawn.
//   - the CHAIN is the animation surface — what a value in transit looks like. It is not a
//     picture of a channel, and not a picture of the messages on one.
//
// So the bead count is not the count of anything on a channel. A chain sits fully populated
// with nothing traversing it; only the lighting moves (and lighting arrives in step 2 — this
// file is geometry only).
//
// NO BEAD POSITION DEPENDS ON ANOTHER BEAD'S POSITION. That is the line separating this from
// the reverted bead-chain wire (memory/project_wire_is_straight_line_not_chain.md), which
// held spacing by neighbour midpoints and therefore followed a drag in O(N²) — measured
// ~1.5s at N≈40. Every offset here is index × spacing along this node's own aim at the
// target: dependency depth 1, which that same memory names as the one escape.

// edgeStepCount is the bead-lattice length of an edge (docs/bead-lattice.md "The count"): ONE
// INTEGER, the number of bead steps between the two nodes' tori. Computed ENTIRELY from lp
// (the SOURCE node's own stored LocalPolar to the target — index arithmetic, no sqrt) plus
// both nodes' kinds:
//
//	N = QuantIR - nodeTorusSteps(srcKind) - nodeTorusSteps(dstKind), minimum 1
//
// PURE INTEGER SUBTRACTION — no division anywhere, not even by a fixed cell count. That used
// to divide QuantIR by a per-bead cell-count constant (4) before subtracting, assuming the node
// lattice's cell was a quarter of a bead step. It was not: the STORED per-entry stepR
// (topology/nodes/<id>/local-polars.json) was 2.0, left over from an earlier, finer lattice,
// while LocalStepR (layout_holder.go) had since become 2.24 — placement read the stored 2.0,
// this division assumed 4 * 2.24, and the two lattices did not nest, so the count
// over-budgeted by ~12% and the surplus beads ran past the target node's tori. The fix
// collapses the two lattices into ONE (LocalStepR now equals BeadStepR — bead_lattice.go's
// BeadStepCells doc comment): QuantIR is already counted in bead-step-sized cells because
// there is no other cell for it to be counted in, so N is exactly QuantIR minus the two
// nodes' own torus extents in the same units, with nothing left to divide.
//
// nodeTorusOuterR is still snapped to a whole number of bead steps (nodeTorusSteps,
// port_geometry.go) rather than measured from width/height, so the subtraction's second and
// third terms are exact integers too — no term here can reintroduce a fraction.
//
// This is a pure function of state a node already owns, called from TWO places that must
// never disagree: chainBeads below (this node's own goroutine, every cycle, for LAYOUT) and
// build.go's allocateWires (load time, for the wire's INITIAL published step count) — both
// call this same function on the same LocalPolar entry, so layout and timing can never read
// two different lengths (the exact divergence docs/bead-lattice.md replaces the old
// arc-length model to close off).
func edgeStepCount(lp wire.LocalPolar, srcKind, dstKind string) int {
	n := lp.QuantIR - nodeTorusSteps(srcKind) - nodeTorusSteps(dstKind)
	if n < 1 {
		return 1
	}
	return n
}

// sendStepsNonBlocking delivers steps to ch, latest-wins: if the buffer already holds
// an undrained stale value (the edgeMover's own goroutine hasn't woken to drain it
// since the last publish), that stale value is dropped and replaced — the same
// "producer sends, one consumer owns its copy" shape speedCh already uses
// (per-goroutine-clock.md "Delivery"), applied to edgeMover.stepsIn. A nil ch (this
// edge had no bound edgeMover.stepsIn, or a bare test nodeMover with no outStepsIn)
// makes every case here select `default`, so this is a silent no-op — this node's own
// goroutine never blocks on it.
func sendStepsNonBlocking(ch chan int, steps int) {
	select {
	case ch <- steps:
		return
	default:
	}
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- steps:
	default:
	}
}

// litBeadIndex maps a bead's progress t (elapsed/ticksToCross, this edge's OWN t) onto the
// index of the chain bead it currently occupies, for a chain of the given STEP count. ok is
// false only when t is outside [0, 1) — off the edge entirely — never because the geometry
// ran out of beads: an index at or past steps is clamped onto the last bead rather than
// reported off-chain.
//
// steps must be the SAME integer chainBeads used to lay the chain out — two different
// lengths for layout vs. lighting is exactly the drift docs/bead-lattice.md's one-integer
// model exists to make impossible (LiveBeadProgress.Steps travels with the bead precisely so
// this can never be re-derived from a second source).
//
//	t*steps = (elapsed/ticksToCross)*steps = elapsed/dwell
//
// which is the same for every edge (dwell is the uniform per-bead constant,
// wire.DwellTicksPerBead), so each index lasts exactly one dwell everywhere.
//
// FLOOR, not round. The lit bead is the last one the traversal has reached, which is what
// floor means; round would instead light the NEAREST, and that ties exactly halfway between
// two beads — not academic here, since two edges reach the same distance via different t
// values and float error would decide the tie differently per edge
// (TestLitBeadIndexSameElapsedLightsSameBead pins this).
func litBeadIndex(t float64, steps int) (int, bool) {
	if t < 0 || t >= 1 || steps <= 0 {
		return 0, false
	}
	// epsilon: t*steps is a float round-trip (t was itself elapsed/ticksToCross), so a
	// bead sitting EXACTLY on bead i's position can land a hair under it and floor to
	// i-1. A bead's own position is a reachable value, not an edge case, so nudge before
	// flooring. 1e-9 against an integer step index is far below anything visible and far
	// above float noise — the same epsilon the retired arc-length version used, kept
	// because the reasoning (a float round-trip, not the unit it multiplies) is unchanged.
	const eps = 1e-9
	idx := int(math.Floor(t*float64(steps) + eps))
	if idx < 0 {
		idx = 0
	}
	if idx >= steps {
		idx = steps - 1
	}
	return idx, true
}

// chainBeads returns THIS node's own placeholder chain beads as node-local offsets, in
// outgoing-edge order (m.outTargets), each edge's beads ordered outward from this node. It
// also PUBLISHES each edge's freshly computed step count onto that edge's own *wire.Out
// (docs/bead-lattice.md "Ownership": the source node owns the count) — the same call that
// lays the chain out on that integer, so the wire's own timing budget and this chain's
// layout can never disagree.
//
// Reads only state this node owns: its own kind/radius and its own stored LOCAL POLAR to
// each neighbour (m.layoutHolderFn().LocalPolarsSnapshot() — the same accessor
// neighborSetCRequantize/armDragAnchor use), never another goroutine's live position. There
// is no cross-goroutine read here, which is why this can run on the emit path.
//
// NO SQRT ANYWHERE in this path (guard: tools/check-no-sqrt-in-chain-beads.sh). A neighbour's
// distance and direction come from this node's OWN stored abc-indices (QuantITheta/
// QuantIPhi/QuantIR) × step constants — index arithmetic, per
// memory/feedback_abc_times_constant_not_rederive.md — not from a cartesian offset's vector
// length or its normalized direction (each internally a sqrt). The only trig is the one
// cartesian↔polar boundary
// conversion per bead (polar2cart), matching the "abc × constant, trig only at the boundary"
// model the demo (docs/demos/polar-drag-3d.html) exists to enforce. edgeStepCount's integer
// subtraction is plain arithmetic, not sqrt, so publishing the count alongside layout does
// not reintroduce one.
//
// Offsets are NODE-LOCAL on purpose: this node moving does not change a single one of them,
// so a move costs one center write instead of degree × N bead positions. Only a NEIGHBOUR
// moving re-aims a chain, and that arrives as the one-hop center message that already
// exists (which is what keeps this node's own LocalPolar to that neighbour current). That is
// the whole constant-time claim.
//
// A target with no stored local polar (never linked, or this node has no LayoutHolder — bare
// movers built directly in tests without one) contributes NO beads rather than beads at a
// made-up direction.
//
// The returned `lit`/`litVal` slices are parallel to the offsets: 1 on the bead each
// in-flight traversal has reached (with that traversal's bead VALUE alongside), 0 elsewhere. A chain with nothing traversing it is fully populated
// and entirely unlit — that resting state is normal, not an absence of data.
//
// The returned `radius` slice is also parallel to the offsets: each bead's own SPHERE
// radius (see the per-edge sizing derivation below `beadOuterR`/`sphereR`) — beads on
// different edges are different sizes on purpose, so this cannot be the shared constant
// wire.BeadRadius the way it used to be.
func (m *nodeMover) chainBeads() (ox, oy, oz []float32, lit []uint8, litVal []int32, radius []float32) {
	if len(m.outTargets) == 0 || m.layoutHolderFn == nil {
		return nil, nil, nil, nil, nil, nil
	}
	lh := m.layoutHolderFn()
	if lh == nil {
		return nil, nil, nil, nil, nil, nil
	}
	localPolars := lh.LocalPolarsSnapshot()
	pole := dir(lh.Pole())
	// Read the clock only when there is a wire to ask about — m.clk is nil in tests that
	// build a bare nodeMover directly (the same convention resolveDest/commitLocal state),
	// and such a mover has no outWires either, so geometry stays testable without a clock.
	var tick int64
	if len(m.outWires) > 0 {
		tick = m.clk.Tick()
	}
	selfTorusR := nodeTorusOuterR(m.geom.Kind)
	// selfCenter is THIS node's own live world center, read the same way
	// emitGeometry/edgeSegment do (nodeWorldPos(m.geom)) — this node's own goroutine
	// is the sole writer of m.geom (applyCenter), so this is a same-goroutine read of
	// state already owned here, not a second cross-goroutine touch.
	selfCenter := nodeWorldPos(m.geom)
	for _, to := range m.outTargets {
		var lp wire.LocalPolar
		found := false
		for _, cand := range localPolars {
			if cand.To == to {
				lp = cand
				found = true
				break
			}
		}
		if !found {
			continue
		}
		// Direction fallback ONLY: this node's own stored, QUANTIZED bearing to the
		// neighbour, re-expressed from its abc-indices about this node's own measurement
		// pole — the same fromAxisFrame call quantized_move.go's requantizePoleTraced and
		// loader_layout.go's reload path use to reconstruct an unchanged neighbour's world
		// direction. Used below ONLY when this node has no live cartesian copy of the
		// neighbour's center yet (m.partnerCenters miss — never linked, or a bare test
		// mover with no pushes) and cannot measure the real bearing; the stored indices are
		// 1-degree angular cells (localStepTheta/Phi, layout_holder.go), so this can point
		// up to half a degree away from where the neighbour actually sits — that residue is
		// the "chain lands beside the target's surface" defect the live measurement below
		// exists to close. QuantITheta/QuantIPhi are still the authoritative RECORDED
		// position (reload reconstruction, requantize read them) — only this CHAIN, a
		// picture of where things are RIGHT NOW, stops reading them once a live center is
		// available; the stored indices themselves are untouched.
		stepTheta, stepPhi, _ := lp.EffectiveSteps()
		ndir := fromAxisFrame(pole, float64(lp.QuantITheta)*stepTheta, float64(lp.QuantIPhi)*stepPhi)

		// The ONE authoritative length: docs/bead-lattice.md's edgeStepCount, computed from
		// this node's own stored LocalPolar and both kinds — no arc, no chord, nothing
		// measured against a cartesian center. Both the placement loop below and this edge's
		// wire (via PublishSteps/outStepsIn just after) read this SAME integer, so layout and
		// timing cannot disagree.
		count := edgeStepCount(lp, m.geom.Kind, m.cascadeKinds[to])

		// Publish this edge's freshly computed step count onto its own *wire.Out
		// (docs/bead-lattice.md "Ownership") and onto its edgeMover's stepsIn (so a live
		// in-flight bead's remaining travel — edgeMover.recomputeGeometry's
		// ReviseInFlightGeometry call — is revised against the same integer too; see
		// edge_mover.go's stepsIn doc comment for why a second delivery is needed instead of
		// the edgeMover reading the Out directly). Both are non-blocking, latest-wins sends —
		// this node's own goroutine never waits on either reader.
		for i, wt := range m.outWireTargets {
			if wt != to {
				continue
			}
			if i < len(m.outWireOuts) && m.outWireOuts[i] != nil {
				m.outWireOuts[i].PublishSteps(count)
			}
			if i < len(m.outStepsIn) && m.outStepsIn[i] != nil {
				sendStepsNonBlocking(m.outStepsIn[i], count)
			}
		}

		// Which bead this edge's traversals have reached. Read from THIS node's own
		// outgoing wire for this target, on this node's own goroutine (it is the goroutine
		// that drives that wire — see nodeMover.outWires), so LiveBeadFractions' single-
		// goroutine contract holds and no other goroutine's state is touched.
		//
		// index -> the traversing bead's VALUE. The value travels because the lit bead takes
		// bead 0's or bead 1's own fill: a bare "is lit" flag could not say which.
		litIdx := map[int]int32{}
		for i, wt := range m.outWireTargets {
			if wt != to || m.outWires[i] == nil {
				continue
			}
			for _, p := range m.outWires[i].LiveBeadFractions(tick) {
				// p.Steps is this bead's OWN step count — the geometry its t was computed
				// against, and must be the same value as `count` above (both trace back to
				// this edge's PublishSteps); passed straight to litBeadIndex rather than
				// re-deriving, so lighting and layout can never read two different lengths.
				if idx, ok := litBeadIndex(p.T, p.Steps); ok {
					litIdx[idx] = int32(p.Val)
				}
			}
		}
		// spacing is the center-to-center distance between consecutive beads on THIS
		// edge, and beadOuterR is this edge's own bead OUTER radius — both default to
		// the lattice's fixed constants (wire.BeadStepR / wire.BeadTorusOuterR), used
		// as a fallback when this node has no live center for `to` yet — e.g. a
		// neighbour that has never pushed an applyCenter, or a bare test nodeMover
		// with no partnerCenters at all — but are normally REPLACED below by the
		// per-edge values that make every bead on this chain touch its neighbour
		// EXACTLY, ends included, with the chain staying a straight line.
		//
		// PER-EDGE BEAD SIZE (the fix; supersedes an earlier "stretch spacing to
		// absorb the residue" attempt that is no longer here). count comes from
		// edgeStepCount, which rounds the SOURCE node's stored, quantized LocalPolar
		// distance to an integer step count (QuantIR) — that rounding is exact by
		// construction of the lattice, but the two nodes' LIVE cartesian positions
		// (what the renderer actually draws them at, nodeWorldPos) are continuous and
		// only coincidentally land on a whole multiple of BeadStepR. So
		// count*BeadStepR (or any FIXED bead size at all) is off from the real
		// surface-to-surface gap by up to half a bead. Stretching `spacing` to absorb
		// that residue (the earlier approach) kept the two ends tangent but left every
		// INTERIOR bead not-quite-touching its neighbour — visibly gappy on short
		// edges. Node positions cannot be snapped to fix this either: a node with 3+
		// neighbours is over-constrained, each edge would want a different snap. The
		// one remaining free parameter is bead SIZE, and it is free PER EDGE — no
		// other edge shares this one's beads, so sizing this edge's beads to its own
		// gap costs nothing elsewhere.
		//
		// Read the ACTUAL gap from this node's own live center and its live copy of
		// the neighbour's center (m.partnerCenters[to], kept current by every
		// applyCenter push — the SAME nodeWorldPos value the renderer streams for both
		// ends, not the stored LocalPolar), then size beads so `count` of them exactly
		// tile it, touching:
		//
		//	beadOuterR = gap / (2*count)             // one bead's outer radius
		//	spacing    = 2*beadOuterR                  // adjacent beads touch exactly
		//	d(i)       = selfTorusR + beadOuterR + i*spacing,  i in [0, count)
		//
		//	near edge of bead 0     = selfTorusR + beadOuterR - beadOuterR = selfTorusR
		//	far edge of bead N-1    = selfTorusR + beadOuterR + (N-1)*spacing + beadOuterR
		//	                        = selfTorusR + 2*beadOuterR*N = selfTorusR + gap
		//
		// which is exact tangency at BOTH ends for ANY count, including count==1 (a
		// single bead just becomes exactly the gap's diameter) — so the old `count > 1`
		// guard around the live-gap solve is gone; every count uses this branch
		// whenever a live partner center is available.
		//
		// The bead's own onscreen SPHERE radius (what the renderer scales the
		// instance to) is smaller than beadOuterR by the ring's proportion of the
		// whole bead, the same ratio wire.BeadRadius/wire.BeadTorusOuterR encodes for
		// the fixed-size fallback: sphereR = beadOuterR / (1 + wire.BeadRingTubeRatio).
		//
		// useLiveAim tracks whether spacing/beadOuterR/sphereR (above) AND direction
		// (below) both came from the ONE live measurement (edgeSurfaceGapAndDir) or
		// both fell back to the stored lattice (BeadStepR/BeadTorusOuterR/BeadRadius
		// spacing/size, ndir bearing) — they must move together, never one live and
		// the other stored, or the chain regresses to exactly the bug this fix closes
		// (a length that agrees with the renderer next to a bearing that doesn't).
		spacing := wire.BeadStepR
		beadOuterR := wire.BeadTorusOuterR
		sphereR := wire.BeadRadius
		liveDir := vec3{}
		useLiveAim := false
		if targetCenter, ok := m.partnerCenters[to]; ok {
			targetTorusR := nodeTorusOuterR(m.cascadeKinds[to])
			gap, dirVec, dirOK := edgeSurfaceGapAndDir(selfCenter, targetCenter, selfTorusR, targetTorusR)
			if dirOK {
				bOuter := gap / (2 * float64(count))
				if bOuter < 0 {
					// A degenerate/negative gap (overlapping nodes): clamp rather
					// than fold beads back past this node's own centre. Exact
					// tangency is not achievable in this degenerate case either way.
					bOuter = 0
				}
				beadOuterR = bOuter
				spacing = 2 * bOuter
				sphereR = bOuter / (1 + wire.BeadRingTubeRatio)
				liveDir = dirVec
				useLiveAim = true
			}
		}
		// One coordinate: bead index i. Offset from this node's centre is
		// selfTorusR + beadOuterR + i*spacing (docs/bead-lattice.md "Placement",
		// derivation above). "Beads never inside a node" falls out of this tangency,
		// with no clamp.
		for i := 0; i < count; i++ {
			d := selfTorusR + beadOuterR + float64(i)*spacing
			var p vec3
			if useLiveAim {
				// liveDir is ALREADY a unit cartesian direction (edgeSurfaceGapAndDir's
				// one Normalize() call) — scaling it by d places the bead directly, with
				// no cartesian->polar->cartesian round trip that would exist only to
				// look like the fallback below. That round trip is what the fallback
				// still needs (ndir only carries an angle, not a vector), but here the
				// live measurement already IS a vector, so converting it to polar and
				// back would just reintroduce float error for no reason.
				p = liveDir.Scale(d)
			} else {
				// Fallback: no live center for `to` yet, so the only direction available
				// is the stored quantized bearing (ndir) — an angle, not a vector — and
				// polar2cart is the one legitimate cartesian<->polar boundary conversion
				// (tools/check-no-sqrt-in-chain-beads.sh) to turn it into a placeable
				// offset. R varies per bead by index arithmetic (d above); Theta/Phi are
				// this neighbour's own fixed bearing, reused unchanged for every bead on
				// this edge.
				p = polar2cart(polar{R: d, Theta: ndir.Theta, Phi: ndir.Phi})
			}
			ox = append(ox, float32(p.X))
			oy = append(oy, float32(p.Y))
			oz = append(oz, float32(p.Z))
			v, isLit := litIdx[i]
			var l uint8
			if isLit {
				l = 1
			}
			lit = append(lit, l)
			litVal = append(litVal, v)
			radius = append(radius, float32(sphereR))
		}
	}
	return ox, oy, oz, lit, litVal, radius
}
