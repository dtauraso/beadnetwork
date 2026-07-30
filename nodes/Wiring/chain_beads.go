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
//	N = QuantIR/BeadStepCells - nodeTorusSteps(srcKind) - nodeTorusSteps(dstKind), minimum 1
//
// PURE INTEGER SUBTRACTION — no division by a float step, no math.Round. That used to read
//
//	gap = QuantIR*stepR - nodeTorusOuterR(srcKind) - nodeTorusOuterR(dstKind)
//	N   = round(gap / BeadStepR)
//
// which could be off by up to half a bead step: nodeTorusOuterR was `nodeRadius(kind) *
// (1+ShadingParamNodeRingTubeRatio)`, an arbitrary float NOT on the bead lattice, so gap was
// never an exact multiple of BeadStepR and round() silently absorbed the remainder — the
// promised tangency at the target end was a rounding coincidence, not a guarantee. The fix
// was not to snap QuantIR alone (that still leaves nodeTorusOuterR off-lattice and the
// division still inexact); it is to make EVERY term here an integer count of bead steps, so
// there is nothing left to round: QuantIR is snapped to a multiple of BeadStepCells at every
// write (wire.SnapQuantIR, called from LayoutHolder.SetLocalPolar/LoadLocalPolars — never
// here, a value that can be stored unsnapped is the bug re-entering by another door), and
// nodeTorusOuterR is snapped to a whole number of bead steps (nodeTorusSteps, port_geometry.go)
// rather than measured from width/height. QuantIR/BeadStepCells is therefore always an exact
// integer bead-step count, and N is plain integer subtraction — the off-by-a-fraction bug
// class is unrepresentable, not merely tuned away (memory/feedback_make_bug_class_unrepresentable.md).
//
// This is a pure function of state a node already owns, called from TWO places that must
// never disagree: chainBeads below (this node's own goroutine, every cycle, for LAYOUT) and
// build.go's allocateWires (load time, for the wire's INITIAL published step count) — both
// call this same function on the same LocalPolar entry, so layout and timing can never read
// two different lengths (the exact divergence docs/bead-lattice.md replaces the old
// arc-length model to close off).
func edgeStepCount(lp wire.LocalPolar, srcKind, dstKind string) int {
	n := lp.QuantIR/wire.BeadStepCells - nodeTorusSteps(srcKind) - nodeTorusSteps(dstKind)
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
func (m *nodeMover) chainBeads() (ox, oy, oz []float32, lit []uint8, litVal []int32) {
	if len(m.outTargets) == 0 || m.layoutHolderFn == nil {
		return nil, nil, nil, nil, nil
	}
	lh := m.layoutHolderFn()
	if lh == nil {
		return nil, nil, nil, nil, nil
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
		// Direction: this node's own stored bearing to the neighbour, re-expressed from its
		// abc-indices about this node's own measurement pole — the same fromAxisFrame call
		// quantized_move.go's requantizePoleTraced and loader_layout.go's reload path use to
		// reconstruct an unchanged neighbour's world direction. No cartesian offset is read.
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
		// One coordinate: bead index i. Bead i's torus is tangent to bead i-1's (bead 0's
		// tangent to this node's own torus, bead count-1's tangent to the target's) —
		// docs/bead-lattice.md "Placement": offset from this node's centre is
		// selfTorusR + BeadTorusOuterR + i*BeadStepR. "Beads never inside a node" falls out
		// of this tangency, with no clamp.
		for i := 0; i < count; i++ {
			d := selfTorusR + wire.BeadTorusOuterR + float64(i)*wire.BeadStepR
			// One trig conversion per bead, at the cartesian↔polar boundary — no sqrt: R
			// varies per bead by index arithmetic (d above), Theta/Phi are this
			// neighbour's own fixed bearing (ndir), reused unchanged for every bead on
			// this edge.
			p := polar2cart(polar{R: d, Theta: ndir.Theta, Phi: ndir.Phi})
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
		}
	}
	return ox, oy, oz, lit, litVal
}
