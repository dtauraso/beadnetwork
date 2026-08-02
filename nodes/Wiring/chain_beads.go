package Wiring

import (
	"fmt"
	"math"

	T "github.com/dtauraso/wirefold/Trace"
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
// INTEGER, the number of bead steps between the two nodes' tori. Computed from the LIVE
// measured center-to-center distance (dist), never a stored cache, plus both nodes' kinds:
//
//	K = round(dist / wire.BeadStepR)
//	N = K - nodeTorusSteps(srcKind) - nodeTorusSteps(dstKind), minimum 1
//
// Under bead CRUD (MODEL.md "Moving a node is CRUD on the edge beads that touch it",
// bead_crud.go) a node's placement does NOT guarantee dist lands on an exact integer
// multiple of BeadStepR for every neighbour simultaneously (that guarantee belonged to the
// rejected global bead-cell solver) — round() here is the real discretizer, not a no-op:
// it is what turns a live, generally off-lattice distance into the whole bead-step count
// this edge actually renders.
//
// PURE INTEGER SUBTRACTION once K is known — no division anywhere else, not even by a
// fixed cell count. That used to divide the STORED QuantIR cache by a per-bead cell-count
// constant (4) before subtracting, assuming the node lattice's cell was a quarter of a
// bead step; before that, it read QuantIR straight off a cache that could go stale
// relative to a node's live position (an offset propagated to a neighbor before that
// neighbor's own commit landed, or a load-time value never re-measured after a drag).
// Reading the live distance instead means this can never disagree with where the node
// ACTUALLY is.
//
// nodeTorusOuterR is still snapped to a whole number of bead steps (nodeTorusSteps,
// port_geometry.go) rather than measured from width/height, so the subtraction's second and
// third terms are exact integers too — no term here can reintroduce a fraction.
//
// This is a pure function of state a node already owns, called from TWO places that must
// never disagree: chainBeads below (this node's own goroutine, every cycle, for LAYOUT) and
// build.go's allocateWires (load time, for the wire's INITIAL published step count) — both
// call this same function on the same live center-to-center distance, so layout and timing
// can never read two different lengths (the exact divergence docs/bead-lattice.md replaces
// the old arc-length model to close off).
func edgeStepCount(dist float64, srcKind, dstKind string) int {
	k := int(math.Round(dist / wire.BeadStepR))
	n := k - nodeTorusSteps(srcKind) - nodeTorusSteps(dstKind)
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
// Reads only state this node owns: its own kind/radius and its own live copy of each
// neighbour's world center (m.partnerCenters — pushed by that neighbour's own
// applyCenter), never reaching into another goroutine's state directly. There is no
// cross-goroutine read here, which is why this can run on the emit path.
//
// NO SQRT ANYWHERE in this path except the ONE edgeCenterDistAndDir call per target (guard:
// tools/check-no-sqrt-in-chain-beads.sh) — a neighbour's distance and direction come from
// this single live measurement, reused for both layout and the published step count, never
// re-measured a second time per bead. The only OTHER trig is a boundary conversion, matching
// the "trig only at the boundary" model the demo (docs/demos/polar-drag-3d.html) exists to
// enforce. edgeStepCount's integer subtraction is plain arithmetic, not sqrt, so publishing
// the count alongside layout does not reintroduce one.
//
// Offsets are NODE-LOCAL on purpose: this node moving does not change a single one of them,
// so a move costs one center write instead of degree × N bead positions. Only a NEIGHBOUR
// moving re-aims a chain, and that arrives as the one-hop center message that already
// exists (moveMsgKindNeighborCenter, which is what keeps m.partnerCenters current). That is
// the whole constant-time claim.
//
// A target with no live partner center yet (never linked, or a bare test mover with no
// pushes) contributes NO beads rather than beads at a made-up direction. There is no
// node-node stored bearing to fall back to — MODEL.md's "the polar model": a node's ONE
// polar coordinate is about the scene centre only; a NEIGHBOUR's position is never stored
// as a second coordinate on this node.
//
// The returned `lit`/`litVal` slices are parallel to the offsets: 1 on the bead each
// in-flight traversal has reached (with that traversal's bead VALUE alongside), 0 elsewhere. A chain with nothing traversing it is fully populated
// and entirely unlit — that resting state is normal, not an absence of data.
//
// Bead SIZE is the single, uniform lattice constant wire.BeadRadius everywhere — every
// bead on every edge is the same size (memory/feedback_uniform_pulse_speed.md's sibling
// rule for size: no per-edge knob). There is no per-bead radius column any more: under
// bead CRUD (MODEL.md "Moving a node is CRUD on the edge beads that touch it",
// bead_crud.go) count*wire.BeadStepR (uniform spacing, uniform size) always lands bead 0's
// near edge on this node's own torus by construction of the placement formula below — no
// per-edge size or stretched spacing is needed for that (see the removed
// "per-edge-bead-scale"/"stretch spacing" history below `spacing`'s declaration).
//
// breadcrumbs (the final return value) is DIAGNOSTIC ONLY (task/log-node4-chain-aim): one
// "chain-aim" event per outgoing target, built here but appended to the CALLER's own
// writeStreamFrame events slice rather than sent via a nested m.writeStreamFrame call —
// writeStreamFrame itself invokes chainBeads to build its frame's chain-bead columns, so a
// second writeStreamFrame call from inside here would recurse (and, before this fix, did:
// chainBeads -> writeStreamFrame -> chainBeads -> ... stack overflow).
func (m *nodeMover) chainBeads() (ox, oy, oz []float32, lit []uint8, litVal []int32, tween []uint8, breadcrumbs []wire.RowEvent) {
	if len(m.outTargets) == 0 {
		return nil, nil, nil, nil, nil, nil, nil
	}
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
		// MODEL.md "the polar model": a node has ONE polar vector PER EDGE, pointing to
		// that edge's starting bead — measured live from this node's own center and its
		// neighbour's own center (m.partnerCenters, pushed by that neighbour's own
		// applyCenter — seeded synchronously for every domain neighbour at construction,
		// node_move.go, so this is populated before this node's own goroutine ever runs).
		// There is NO stored node-node bearing record here any more (wire.LocalPolar and
		// its requantize machinery are deleted): a target with no live partner center yet
		// (never linked, or a bare test mover with no pushes) contributes no beads, exactly
		// like the old "no LocalPolar entry" skip.
		targetCenter, haveTargetCenter := m.partnerCenters[to]
		if !haveTargetCenter {
			continue
		}

		// The ONE authoritative length: docs/bead-lattice.md's edgeStepCount, computed
		// from the LIVE center-to-center distance — the model has no stored fallback any
		// more (wire.LocalPolar deleted): a target with no live measurement was already
		// skipped above via haveTargetCenter. dist and liveDir (the DIRECTION, consumed
		// further down) come from the SAME edgeCenterDistAndDir call — one measurement of
		// the edge, not two (that function's own doc comment). Both the placement loop
		// below and this edge's wire (via PublishSteps/outStepsIn just after) read this
		// SAME integer, so layout and timing cannot disagree.
		//
		// edgeCenterDistAndDir's one sqrt-based vector-length/normalize pair is
		// deliberately NOT inlined here: this file is guarded against a cartesian sqrt
		// (tools/check-no-sqrt-in-chain-beads.sh) so bead placement stays a direct read of
		// the live measurement; the sqrt itself lives in port_geometry.go, which already
		// computes edgeSegment the same way.
		dist, liveDir, ok := edgeCenterDistAndDir(selfCenter, targetCenter)
		if !ok {
			continue
		}
		count := edgeStepCount(dist, m.geom.Kind, m.cascadeKinds[to])

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
					// litBeadIndex works on the COARSE lattice (p.Steps is the wire's own
					// step count, untouched by the overlay). On the half-step lattice
					// ordinary bead k sits at ODD index 2k+1 — index 0 is the node-end
					// joint — so the traversal lights the same positions at the same
					// speed, and a tween is never the lit one.
					if m.tweens {
						idx = 2*idx + 1
					}
					litIdx[idx] = int32(p.Val)
				}
			}
		}
		// spacing is the center-to-center distance between consecutive beads on THIS
		// edge, and it is the single UNIFORM lattice constant wire.BeadStepR everywhere
		// (memory/feedback_uniform_pulse_speed.md's sibling rule for size/spacing: no
		// per-edge knob). This used to be stretched per edge (an earlier "absorb the
		// half-bead residue in spacing" fix) and, later, per-edge bead SIZE was made the
		// free parameter instead (commit d50fab83, "each edge sizes its own beads so
		// straight chains touch") — both existed because edgeStepCount's count could
		// disagree with the two nodes' LIVE cartesian gap by up to half a bead, leaving a
		// residue somewhere for spacing or size to absorb.
		//
		// Under bead CRUD (MODEL.md "Moving a node is CRUD on the edge beads that touch
		// it", bead_crud.go) bead 0's near edge is exactly on this node's own torus by
		// construction of the placement formula below, for every count including
		// count==1, regardless of whether dist(node, neighbor) happens to land on an
		// exact multiple of BeadStepR — that global guarantee belonged to the rejected
		// bead-cell solver and is not one this model makes. count (edgeStepCount, read
		// from the same live distance) times the fixed wire.BeadStepR is what fixes the
		// FAR edge's residue against the neighbour's torus instead; it is bounded by
		// round(), never bent into an arc.
		//
		// liveDir (computed above, alongside dist) is the ONLY aim now: the node's own
		// live measurement to its neighbour, no stored fallback bearing (wire.LocalPolar
		// deleted). Direction and spacing/size are independent: spacing and size are
		// always the fixed constants, only the AIM varies.
		//
		// DIAGNOSTIC ONLY (task/log-node4-chain-aim): one breadcrumb per outgoing target
		// per chainBeads() call. Gated on m.tr != nil exactly like emitGeometry's own
		// breadcrumb calls elsewhere in this package — cheap no-op with no stream wired
		// (headless tests, bare movers).
		if m.tr != nil {
			targetRow := int32(-1)
			if m.nodeRowFor != nil {
				if r, ok := m.nodeRowFor(to); ok {
					targetRow = r
				}
			}
			// liveDir is ALREADY a unit vector (Normalize() above), so its own theta/phi
			// come from math.Acos/math.Atan2 directly on the unit components — no second
			// vector-length or re-normalize call, which tools/check-no-sqrt-in-chain-beads.sh
			// bans in this file (trig itself is allowed, only the sqrt-fingerprinted
			// helpers are not).
			liveTheta := math.Acos(clamp(liveDir.Y, -1, 1))
			livePhi := math.Atan2(liveDir.Z, liveDir.X)
			value := fmt.Sprintf(
				"to=%s count=%d K=%d liveDir=(theta=%.4f,phi=%.4f)",
				to, count, int(math.Round(dist/wire.BeadStepR)), liveTheta, livePhi)
			m.tr.Breadcrumb("chain-aim", m.id, to, value)
			breadcrumbs = append(breadcrumbs, wire.RowEvent{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbChainAim, Debug: 1,
				NodeRow: m.nodeRow, PortRow: -1, TargetRow: targetRow, TargetPortRow: -1,
				EdgeRow: -1, Slot: -1, Text: value,
			})
		}
		// One coordinate: bead index i. Offset from this node's centre is
		// selfTorusR + wire.BeadTorusOuterR + i*wire.BeadStepR (docs/bead-lattice.md
		// "Placement"). "Beads never inside a node" falls out of this tangency, with no
		// clamp.
		// TWEEN LATTICE. With the tween overlay on, a JOINT bead sits in every gap along this
		// chain, so the render lattice halves its step: 2N beads instead of N, EVEN index a
		// joint, ODD index ordinary bead (i-1)/2 at exactly the offset it always had. Index 0
		// is the joint between the NODE and bead 0 — that gap gets one too, which is why this
		// is 2N and not 2N-1.
		//
		// The lattice is UNIFORM, including at the node end: index 0 lands half a step before
		// bead 0, which is the node's own torus radius, so a joint there straddles the node's
		// surface — overlapping the node and overlapping bead 0, which is what a joint is for.
		// An earlier version special-cased index 0 to sit fully outside the node; that made it
		// a different size problem at one end for no gain, since a joint is SUPPOSED to cross
		// the boundary it joins.
		//
		// Only the render lattice changes. `count` above is this edge's published STEP count
		// (PublishSteps, the wire's own timing) and is deliberately NOT doubled: doubling it
		// would make the same traversal take twice as many steps and visibly halve the
		// animation speed. Timing stays coarse; only what is drawn gets finer.
		step := wire.BeadStepR
		base := selfTorusR + wire.BeadTorusOuterR
		renderCount := count
		if m.tweens && count > 0 {
			step = wire.BeadStepR / 2
			base = selfTorusR + wire.BeadTorusOuterR - step
			renderCount = 2 * count
		}
		offsetAt := func(i int) float64 {
			return base + float64(i)*step
		}
		// aimUnit is the live direction, carried as a plain unit vector: this is what
		// gets broadcast to this edge's bead-actor chain (reconcileBeadChain,
		// bead_chain.go), which resolves each bead's own position from it directly (one
		// hop, dependency depth 1 — no neighbour read). Bead 0's resolved position IS
		// this node's own "node -> first bead" polar vector (MODEL.md): owned by that
		// bead's own goroutine, never a second stored copy here.
		aimUnit := liveDir
		// Production call site for the bead-actor primitive (nodes/wire/bead_actor.go,
		// bead_wake_group.go): nil in every bare-literal test nodeMover, so this stays a
		// no-op there and chainBeads keeps its pure, synchronous, deterministic contract
		// (see beadTickFn's own doc comment). In production this reconciles this edge's
		// live *wire.Bead goroutine count to `count` and broadcasts fresh geometry when
		// the aim or count changed.
		var actorChain *edgeBeadChain
		if m.beadTickFn != nil {
			actorChain = m.reconcileBeadChain(to, renderCount, offsetAt, aimUnit)
		}
		for i := 0; i < renderCount; i++ {
			var p vec3
			if actorChain != nil && i < len(actorChain.valid) && actorChain.valid[i] {
				// The bead's own goroutine already resolved this position from the
				// broadcast above (or an earlier one at the same aim) — use it rather
				// than recomputing the identical value here.
				p = actorChain.last[i].Position
			} else {
				// liveDir is ALREADY a unit cartesian direction — scaling it by d places
				// the bead directly, with no cartesian->polar->cartesian round trip.
				p = liveDir.Scale(offsetAt(i))
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
			// EVEN index on the half-step lattice is a joint (index 0 is the node-end one);
			// odd index is an ordinary bead. With the overlay off every index is ordinary.
			// Emitted per row because parity is not recoverable downstream — see
			// bufLayoutChainBead.Tween.
			var tw uint8
			if m.tweens && i%2 == 0 {
				tw = 1
			}
			tween = append(tween, tw)
		}
	}
	return ox, oy, oz, lit, litVal, tween, breadcrumbs
}
