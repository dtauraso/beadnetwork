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

// beadCount is how many touching beads fit along an arc of the given length, at constant
// chainBeadSpacing. Both the placement loop and the lit-index quantisation work in this ONE
// coordinate — distance travelled d along the arc, d in [0, arc] — so they cannot disagree
// about how long the chain is. That disagreement (arc-length lighting against a
// center-distance-minus-radii layout) was the bug: the tail of the chain sat past the
// largest distance litBeadIndex could ever produce, so the last 1-2 beads of every edge were
// never reachable. See docs/beads-are-the-edge.md and the arc-source doc comment on
// chainBeads below.
//
// count = floor(arc / spacing): the number of FULL spacing-wide intervals that fit inside
// [0, arc]. FLOOR, not ceil — this is what keeps the last bead from ever overshooting past
// the target's surface (see chainBeads' offset comment): with count*spacing <= arc always
// true by construction, the last bead's far edge (at selfR + count*spacing) never passes
// the target surface (at selfR + arc), landing exactly tangent when arc is an exact multiple
// of spacing and strictly outside otherwise. A ceil-rounded count was tried and rejected: for
// an arc that is not an exact multiple, ceil admits a final PARTIAL interval whose bead can
// overshoot the target surface by up to a full spacing (nearly one whole bead diameter) —
// the very "buried in the target" defect this design exists to rule out.
//
// Every index in [0, count-1) is still reachable by some d < arc — litBeadIndex's floor over
// [0, count*spacing) covers all of them before d ever reaches the (< one spacing) leftover
// tail beyond count*spacing; that tail floors to `count` and is clamped onto the last bead
// (index count-1), which is why litBeadIndex clamps rather than reporting !ok there. No bead
// is ever unreachable, and the FIRST bead is always reachable at d=0. See
// docs/beads-are-the-edge.md and the arc-source doc comment on chainBeads below.
func beadCount(arc float64) int {
	if arc <= 0 {
		return 0
	}
	return int(arc / chainBeadSpacing)
}

// litBeadIndex maps a bead's progress t (elapsed/ticksToCross, this edge's OWN t) onto the
// index of the chain bead it currently occupies, for a chain of the given arc. ok is false
// only when t is outside [0, 1) — off the edge entirely — never because the geometry ran out
// of beads: an index at or past beadCount(arc) is clamped onto the last bead rather than
// reported off-chain (see beadCount's doc comment for why that tail exists).
//
// arc must be the SAME arc beadCount/chainBeads used to lay the chain out — two different
// lengths for layout vs. lighting is exactly the drift that caused the unreachable-tail bug
// this replaces (docs/beads-are-the-edge.md).
//
//	t*arc = (elapsed/ticksToCross)*arc = elapsed*pulseSpeed
//
// which is the same for every edge, so each index lasts exactly chainBeadSpacing/pulseSpeed
// ticks everywhere — the constant dwell docs/beads-are-the-edge.md rests on.
//
// FLOOR, not round. The lit bead is the last one the traversal has reached, which is what
// floor means; round would instead light the NEAREST, and that ties exactly halfway between
// two beads — not academic here, since two edges reach the same distance via different t
// values and float error would decide the tie differently per edge
// (TestLitBeadIndexSameElapsedLightsSameBead pins this).
func litBeadIndex(t, arc float64) (int, bool) {
	if t < 0 || t >= 1 || arc <= 0 {
		return 0, false
	}
	// epsilon: t*arc is a float round-trip (t was itself elapsed/ticksToCross), so a bead
	// sitting EXACTLY on bead i's position can land a hair under it and floor to i-1. A
	// bead's own position is a reachable value, not an edge case, so nudge before flooring.
	// 1e-9 against a spacing of 8 world units is far below anything visible and far above
	// float noise.
	const eps = 1e-9
	idx := int(math.Floor((t*arc + eps) / chainBeadSpacing))
	if idx < 0 {
		return 0, false
	}
	if n := beadCount(arc); idx >= n {
		idx = n - 1
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
		selfR := nodeRadius(m.geom.Kind)

		// arc is the ONE length both the layout below and the lighting above must agree
		// on — divergence between them is exactly what caused the unreachable-tail bug
		// (docs/beads-are-the-edge.md): litBeadIndex used to multiply t by the wire's
		// PUBLISHED port-to-port ArcLength while the chain was laid out to a DIFFERENT,
		// center-distance-minus-radii length, so the chain's tail sat past the largest
		// distance the lit index could ever reach.
		//
		// Preferred source: this edge's own *wire.Out, bound alongside outWires in
		// moverRegistry.bind (outWireOuts, node_mover.go) — its Geom().ArcLength is the
		// authoritative port-to-port arc PublishGeom sets from the loaded/edited
		// geometry, the same value the wire's own t = elapsed/ticksToCross was computed
		// against.
		//
		// Fallback (required, not optional): ArcLength is 0 before geometry is
		// published (early startup) and bare nodeMovers built directly in tests have no
		// Out at all (outWireOuts is nil then). In both cases fall back to the local
		// surface-to-surface estimate — all node-local data (this node's own kind, the
		// target's kind from cascadeKinds, and the center distance already computed
		// above) — so a chain still lays out before geometry exists or under test. The
		// published arc wins whenever it is available and nonzero.
		arc := 0.0
		for i, wt := range m.outWireTargets {
			if wt != to || i >= len(m.outWireOuts) || m.outWireOuts[i] == nil {
				continue
			}
			if a := m.outWireOuts[i].Geom().ArcLength; a > 0 {
				arc = a
				break
			}
		}
		if arc <= 0 {
			arc = length - selfR - nodeRadius(m.cascadeKinds[to])
		}
		count := beadCount(arc)
		if count == 0 {
			// Either the estimate collapsed to <=0 (nodes close enough that no bead
			// fits in the surface-to-surface gap) or the published arc is somehow
			// nonpositive. Emit none rather than beads buried in one node or the other.
			continue
		}
		dir := offset.Normalize()
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
				// p.Arc is this bead's OWN arc — the geometry its t was computed
				// against — and must be the same value as `arc` above (both come from
				// the same Out.Geom().ArcLength once published); passed straight to
				// litBeadIndex rather than re-deriving, so lighting and layout can never
				// read two different lengths again.
				if idx, ok := litBeadIndex(p.T, p.Arc); ok {
					litIdx[idx] = int32(p.Val)
				}
			}
		}
		// One coordinate: distance travelled d along the arc. Bead i occupies
		// [i*spacing, (i+1)*spacing), so its drawn offset from this node's centre is the
		// CENTRE of that interval, measured out from this node's own surface:
		// selfR + spacing/2 + i*spacing. chainBeadSpacing/2 == ShadingParamBeadRadius (a
		// bead's own radius), so bead 0's near edge lands exactly tangent outside this
		// node's surface, and — because count = floor(arc/spacing) keeps count*spacing <=
		// arc always (beadCount's doc comment) — the last bead's far edge (selfR +
		// count*spacing) never passes the target's surface (selfR + arc), landing exactly
		// tangent inside it when arc is an exact multiple of spacing and strictly clear of
		// it otherwise. "Beads never inside a node" therefore falls out of this geometry,
		// not out of a separate startAt/endAt clamp.
		for i := 0; i < count; i++ {
			d := selfR + chainBeadSpacing/2 + float64(i)*chainBeadSpacing
			p := dir.Scale(d)
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
