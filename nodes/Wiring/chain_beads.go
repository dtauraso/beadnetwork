package Wiring

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

// chainBeadSpacing is the CONSTANT world distance between adjacent chain beads. Constant
// spacing is what makes uniform pulse speed structural rather than computed
// (memory/feedback_uniform_pulse_speed.md): with the count proportional to length
// (count = len/spacing), a constant dwell per bead gives
//
//	total = count × dwell = (len/spacing) × dwell = len / (spacing/dwell)
//
// which is exactly today's ticksToCross = arcLength / pulseSpeed. So there is no per-edge
// arc-length division to do — a long chain simply has more beads, each dwelt on for the same
// time. See docs/beads-are-the-edge.md open question 4.
//
// It is a chosen constant, not derived from anything: it sets how finely a traversal is
// quantised visually. Stated outright rather than dressed up as a derivation.
const chainBeadSpacing = 12.0

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
// The returned `lit` slice is parallel to the offsets: 1 on the bead each in-flight
// traversal has reached, 0 elsewhere. A chain with nothing traversing it is fully populated
// and entirely unlit — that resting state is normal, not an absence of data.
func (m *nodeMover) chainBeads() (ox, oy, oz []float32, lit []uint8) {
	if len(m.outTargets) == 0 {
		return nil, nil, nil, nil
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
		if length < chainBeadSpacing {
			// Closer than one spacing: no interior bead position exists. Emitting a single
			// bead here would put it past the target.
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
		litIdx := map[int]bool{}
		for i, target := range m.outWireTargets {
			if target != to || m.outWires[i] == nil {
				continue
			}
			for _, t := range m.outWires[i].LiveBeadFractions(tick) {
				litIdx[int(t*float64(int(length/chainBeadSpacing)))] = true
			}
		}
		// count = len/spacing, the length-proportional count the constant-spacing argument
		// above depends on. Beads sit at index × spacing OUTWARD FROM THIS NODE, starting at
		// index 1 — index 0 would be this node's own center.
		count := int(length / chainBeadSpacing)
		for i := 1; i <= count; i++ {
			p := dir.Scale(float64(i) * chainBeadSpacing)
			ox = append(ox, float32(p.X))
			oy = append(oy, float32(p.Y))
			oz = append(oz, float32(p.Z))
			var l uint8
			if litIdx[i] {
				l = 1
			}
			lit = append(lit, l)
		}
	}
	return ox, oy, oz, lit
}
