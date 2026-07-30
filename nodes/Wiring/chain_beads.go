package Wiring

import "math"

// chain_beads.go — the node-owned placeholder bead chain that IS the visual of an edge.
// Design and staging: docs/beads-are-the-edge.md.
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

// chainBeadSpacing is the CONSTANT world distance between adjacent chain beads: exactly one
// bead DIAMETER, so adjacent beads TOUCH and a chain reads as a solid line with no gaps.
// Derived from ShadingParamBeadRadius (shading_params.go), the same constant the renderer
// sizes a bead from — so the "no gaps" property cannot drift into a gap by one side changing
// its own copy of the radius.
//
// Constant spacing is also what makes uniform pulse speed structural rather than computed
// (memory/feedback_uniform_pulse_speed.md): with the count proportional to length
// (count = len/spacing), a constant dwell per bead gives
//
//	total = count × dwell = (len/spacing) × dwell = len / (spacing/dwell)
//
// which is exactly today's ticksToCross = arcLength / pulseSpeed. So there is no per-edge
// arc-length division to do — a long chain simply has more beads, each dwelt on for the same
// time. See docs/beads-are-the-edge.md open question 4.
//
// How finely a traversal is quantised visually therefore follows from the bead SIZE rather
// than being its own knob: touching beads is the spec, so the step is a diameter.
const chainBeadSpacing = 2 * ShadingParamBeadRadius

// beadsInSpan is how many touching beads fit in the surface-to-surface span [startAt, endAt],
// at constant chainBeadSpacing. Both the placement loop and the lit-index quantisation call
// it, so they cannot disagree about how long the chain is — the bug that would put the lit
// index past the last bead.
//
// CONSEQUENCE worth stating: the span is center distance MINUS both node radii and two bead
// radii, so it is not exactly proportional to arc length, and the lit bead's world speed
// therefore varies slightly between pairs of different node KINDS (radius 15 vs radius 9
// nodes swallow different amounts). memory/feedback_uniform_pulse_speed.md's uniform speed
// holds for the wire's own ticksToCross, which is unchanged; it is the VISIBLE span that is
// shortened. The old moving bead ran port-to-port and was shortened the same way, so this is
// not a new deviation — but it is a real one, and it is not the constant-spacing derivation's
// doing.
func beadsInSpan(startAt, endAt float64) int {
	if endAt < startAt {
		return 0
	}
	return int((endAt-startAt)/chainBeadSpacing) + 1
}

// litBeadIndex maps a bead's progress onto the index of the chain bead it currently occupies,
// for a chain spanning [startAt, endAt]. ok is false when the bead is not over the chain (before
// the first bead or past the last), in which case nothing is lit.
//
// It converts to DISTANCE COVERED first, and the length it multiplies by must be the bead's OWN
// ARC — the geometry its t was computed against. Then
//
//	t*arc = (elapsed/ticksToCross)*arc = elapsed*pulseSpeed
//
// which is the same for every edge, so each index lasts exactly chainBeadSpacing/pulseSpeed
// ticks everywhere. That is the constant dwell docs/beads-are-the-edge.md rests on: two beads
// placed in one emission advance bead-for-bead whatever the edge lengths, and a longer edge
// simply has further to go.
//
// Two versions of this were wrong in the same way, each with a length that was not the arc:
//
//   - int(t * beadCount) — t climbs faster on a shorter edge, so equal counts stepped at
//     unequal rates. Node 1's edges differ 1.9% in length while BOTH chains hold 28 beads.
//   - t * centerDistance — off by (centerDistance/arc), which differs per edge because the arc
//     is port-to-port geometry and the center separation is not.
//
// Both read on screen as one bead permanently ahead of the other.
//
// FLOOR, not round. The lit bead is the last one the traversal has reached, which is what floor
// means; round would instead light the NEAREST, and that ties exactly halfway between two beads.
// A tie is not academic here: the two edges reach the same distance via different t values, so
// float error decides the tie differently per edge and the two chains disagree by a bead at every
// midpoint. A test asserts the two edges agree at equal distance and it caught exactly that.
func litBeadIndex(t, arc, startAt, endAt float64) (int, bool) {
	// epsilon: t*length is a float round-trip (t was itself elapsed/ticksToCross), so a bead
	// sitting EXACTLY on bead i's position can land a hair under it and floor to i-1. A bead's
	// own position is a reachable value, not an edge case, so nudge before flooring. 1e-9 against
	// a spacing of 8 world units is far below anything visible and far above float noise.
	const eps = 1e-9
	idx := int(math.Floor((t*arc - startAt + eps) / chainBeadSpacing))
	if idx < 0 || idx >= beadsInSpan(startAt, endAt) {
		return 0, false
	}
	return idx, true
}

// chainBeads returns THIS node's own placeholder chain beads as node-local offsets, in
// outgoing-edge order (m.outTargets), each edge's beads ordered outward from this node.
//
// Reads only state this node owns: its own center (m.geom) and its own partnerCenters map —
// each neighbour's last-known center, "written ONLY by this node's own goroutine" and kept
// current by that neighbour's own applyCenter push (nodeMover.partnerCenters' doc comment).
// There is no cross-goroutine read of another node's live position here, which is why this
// can run on the emit path.
//
// Offsets are NODE-LOCAL on purpose: this node moving does not change a single one of them,
// so a move costs one center write instead of degree × N bead positions. Only a NEIGHBOUR
// moving re-aims a chain, and that arrives as the one-hop center message that already
// exists. That is the whole constant-time claim.
//
// A target with no known center, or one sitting on top of this node (zero-length offset, no
// defined direction), contributes NO beads rather than beads at a made-up direction.
//
// The returned `lit`/`litVal` slices are parallel to the offsets: 1 on the bead each
// in-flight traversal has reached (with that traversal's bead VALUE alongside), 0 elsewhere. A chain with nothing traversing it is fully populated
// and entirely unlit — that resting state is normal, not an absence of data.
func (m *nodeMover) chainBeads() (ox, oy, oz []float32, lit []uint8, litVal []int32) {
	if len(m.outTargets) == 0 {
		return nil, nil, nil, nil, nil
	}
	self := nodeWorldPos(m.geom)
	// Read the clock only when there is a wire to ask about — m.clk is nil in tests that
	// build a bare nodeMover directly (the same convention resolveDest/commitLocal state),
	// and such a mover has no outWires either, so geometry stays testable without a clock.
	var tick int64
	if len(m.outWires) > 0 {
		tick = m.clk.Tick()
	}
	for _, to := range m.outTargets {
		target, ok := m.partnerCenters[to]
		if !ok {
			continue
		}
		offset := target.Sub(self)
		length := offset.Length()
		// The chain runs between the two node SURFACES, not between their centers. A bead
		// placed by distance-from-center alone sits INSIDE a node whenever that distance is
		// under the node's radius — which it was at BOTH ends: the first bead (one spacing =
		// 8 out) fell inside a radius-15 node, and the count ran to the target's center so
		// the last beads were inside the target.
		//
		// Both radii are already node-local data: this node's own kind, and the target's kind
		// from m.cascadeKinds (stored per node in cascade-edges.json). Nothing new has to be
		// fetched or messaged for this.
		//
		// A bead's own radius is added at each end too, so the end beads sit TANGENT outside
		// the spheres rather than half-buried in them.
		startAt := nodeRadius(m.geom.Kind) + ShadingParamBeadRadius
		endAt := length - nodeRadius(m.cascadeKinds[to]) - ShadingParamBeadRadius
		if endAt < startAt {
			// The two nodes are close enough that no bead fits in the gap between their
			// surfaces. Emit none rather than beads buried in one node or the other.
			continue
		}
		dir := offset.Normalize()
		// Which bead this edge's traversals have reached. Read from THIS node's own
		// outgoing wire for this target, on this node's own goroutine (it is the goroutine
		// that drives that wire — see nodeMover.outWires), so LiveBeadFractions' single-
		// goroutine contract holds and no other goroutine's state is touched.
		//
		// t is the only thing the animation needs: the chain is fixed, so "where has this
		// traversal got to" is one number. index = t × count is that number quantised onto
		// the beads that already exist — arithmetic on an index, not a re-derived position
		// (memory/feedback_abc_times_constant_not_rederive.md).
		// index -> the traversing bead's VALUE. The value travels because the lit bead takes
		// bead 0's or bead 1's own fill: a bare "is lit" flag could not say which.
		litIdx := map[int]int32{}
		for i, target := range m.outWireTargets {
			if target != to || m.outWires[i] == nil {
				continue
			}
			for _, p := range m.outWires[i].LiveBeadFractions(tick) {
				// Quantised onto the beads that actually exist — the surface-to-surface span,
				// not the center-to-center length, or the lit index would run off the end of
				// the chain by however many beads the two node radii swallow.
				// From DISTANCE COVERED, not from the fraction. p.T is elapsed /
				// ticksToCross, and ticksToCross = arcLength / pulseSpeed, so t climbs
				// FASTER on a shorter edge. Quantising t onto the bead count therefore
				// steps two chains at different rates: node 1's edges measure 259.2 and
				// 254.3, a 1.9% difference, which drifts the two lit indices apart by up
				// to half a bead and reads as a permanent one-bead offset between the two
				// animations. Both chains have the same bead count, so the count was never
				// the problem — the RATE was.
				//
				// p.T*length is world distance from this node's center; subtracting startAt
				// and dividing by the spacing gives the bead index. Each index then lasts
				// exactly spacing/pulseSpeed ticks on EVERY edge, which is the constant
				// dwell the design rests on (docs/beads-are-the-edge.md): two beads placed
				// in one emission advance bead-for-bead regardless of edge length, and a
				// longer edge simply has further to go.
				//
				// p.Arc, NOT this edge's center separation: only the bead's own arc turns t
				// back into the distance it has actually covered.
				if idx, ok := litBeadIndex(p.T, p.Arc, startAt, endAt); ok {
					litIdx[idx] = int32(p.Val)
				}
			}
		}
		// Length-proportional count, still at constant spacing — the property the uniform-speed
		// argument above rests on. Index 0 is now the FIRST bead outside this node's surface,
		// not this node's center.
		count := beadsInSpan(startAt, endAt)
		for i := 0; i < count; i++ {
			p := dir.Scale(startAt + float64(i)*chainBeadSpacing)
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
