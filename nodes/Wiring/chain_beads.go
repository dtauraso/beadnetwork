package Wiring

import (
	"os"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/beadindex"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/stepdeliver"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// chainAimTraceEnabled gates the per-tick "chain-aim" breadcrumb below. Read ONCE at
// process start, before any goroutine exists — the same "one env var, read once" shape
// wire.edgeBeadTraceEnabled uses for the other high-volume trace, so no goroutine ever
// races on it and no synchronization is needed.
var chainAimTraceEnabled = os.Getenv("WIREFOLD_CHAIN_AIM_TRACE") == "1"

// chain_beads.go — the node-owned placeholder bead chain that IS the visual of an edge.
// Design and staging: docs/bead-model/beads-are-the-edge.md. The LENGTH model (one integer bead-step
// count, tangent placement, no arc) is docs/bead-model/bead-lattice.md — this file implements it.
// The length itself (edgeStepCount) is in chain_length.go, the step-count publish helper
// (stepdeliver.SendStepsNonBlocking) is in nodes/Wiring/stepdeliver, and the
// progress->index math (beadindex.LitBeadIndex, used by tests and documented here for the
// lit-pulse placement below) is in nodes/Wiring/beadindex — this file keeps the one thing
// that reads all three: the placement loop itself.
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

// chainBeads returns THIS node's own placeholder chain beads as node-local offsets, in
// outgoing-edge order (m.outs.outTargets), each edge's beads ordered outward from this node. It
// also PUBLISHES each edge's freshly computed step count onto that edge's own *wire.Out
// (docs/bead-model/bead-lattice.md "Ownership": the source node owns the count) — the same call that
// lays the chain out on that integer, so the wire's own timing budget and this chain's
// layout can never disagree.
//
// Reads only state this node owns: its own kind/radius and its own live copy of each
// neighbour's world center (m.topo.partnerCenters — pushed by that neighbour's own
// applyCenter), never reaching into another goroutine's state directly. There is no
// cross-goroutine read here, which is why this can run on the emit path.
//
// NO SQRT ANYWHERE in this path except the ONE nodegeom.EdgeCenterDistAndDir call per target (guard:
// tools/network/beads/check-no-sqrt-in-chain-beads.sh) — a neighbour's distance and direction come from
// this single live measurement, reused for both layout and the published step count, never
// re-measured a second time per bead. The only OTHER trig is a boundary conversion, matching
// the "trig only at the boundary" model this file enforces. edgeStepCount's integer subtraction is plain arithmetic, not sqrt, so publishing
// the count alongside layout does not reintroduce one.
//
// Offsets are NODE-LOCAL on purpose: this node moving does not change a single one of them,
// so a move costs one center write instead of degree × N bead positions. Only a NEIGHBOUR
// moving re-aims a chain, and that arrives as the one-hop center message that already
// exists (movemsg.KindNeighborCenter, which is what keeps m.topo.partnerCenters current). That is
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
// Bead SIZE is the single, uniform lattice constant lattice.BeadRadius everywhere — every
// bead on every edge is the same size (memory/feedback_uniform_pulse_speed.md's sibling
// rule for size: no per-edge knob). There is no per-bead radius column any more: under
// bead CRUD (MODEL.md "Moving a node is CRUD on the edge beads that touch it",
// bead_crud.go) count*lattice.BeadStepR (uniform spacing, uniform size) always lands bead 0's
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
func (m *nodeGeometry) chainBeads() (ox, oy, oz []float32, lit []uint8, litVal []int32, breadcrumbs []wire.RowEvent) {
	if len(m.outs.outTargets) == 0 {
		return nil, nil, nil, nil, nil, nil
	}
	// Read the clock only when there is a wire to ask about — m.clocks.clk is nil in tests that
	// build a bare nodeMover directly (the same convention resolveDest/commitLocal state),
	// and such a mover has no outWires either, so geometry stays testable without a clock.
	var tick int64
	if len(m.outs.outWires) > 0 {
		tick = m.clocks.clk.Tick()
	}
	selfTorusR := nodegeom.NodeTorusOuterR(m.geom.Kind)
	// selfCenter is THIS node's own live world center, read the same way
	// emitGeometry/EdgeSegment do (nodegeom.NodeWorldPos(m.geom)) — this node's own goroutine
	// is the sole writer of m.geom (applyCenter), so this is a same-goroutine read of
	// state already owned here, not a second cross-goroutine touch.
	selfCenter := nodegeom.NodeWorldPos(m.geom)
	for _, to := range m.outs.outTargets {
		// MODEL.md "the polar model": a node has ONE polar vector PER EDGE, pointing to
		// that edge's starting bead — measured live from this node's own center and its
		// neighbour's own center (m.topo.partnerCenters, pushed by that neighbour's own
		// applyCenter — seeded synchronously for every domain neighbour at construction,
		// move_dispatch_construct.go, so this is populated before this node's own goroutine ever runs).
		// There is NO stored node-node bearing record here any more (wire.LocalPolar and
		// its requantize machinery are deleted): a target with no live partner center yet
		// (never linked, or a bare test mover with no pushes) contributes no beads, exactly
		// like the old "no LocalPolar entry" skip.
		targetCenter, haveTargetCenter := m.topo.partnerCenters[to]
		if !haveTargetCenter {
			continue
		}

		// The ONE authoritative length: docs/bead-model/bead-lattice.md's edgeStepCount, computed
		// from the LIVE center-to-center distance — the model has no stored fallback any
		// more (wire.LocalPolar deleted): a target with no live measurement was already
		// skipped above via haveTargetCenter. dist and liveDir (the DIRECTION, consumed
		// further down) come from the SAME nodegeom.EdgeCenterDistAndDir call — one measurement of
		// the edge, not two (that function's own doc comment). Both the placement loop
		// below and this edge's wire (via PublishSteps/outStepsIn just after) read this
		// SAME integer, so layout and timing cannot disagree.
		//
		// nodegeom.EdgeCenterDistAndDir's one sqrt-based vector-length/normalize pair is
		// deliberately NOT inlined here: this file is guarded against a cartesian sqrt
		// (tools/network/beads/check-no-sqrt-in-chain-beads.sh) so bead placement stays a direct read of
		// the live measurement; the sqrt itself lives in port_geometry.go, which already
		// computes nodegeom.EdgeSegment the same way.
		// ChainEdgeGeometry is nodegeom.EdgeCenterDistAndDir + nodegeom.EdgeStepCount,
		// unchanged, named as a phase (beadindex/chain_edge_layout.go): the ONE live
		// measurement of this edge and the ONE integer bead-step count derived from it,
		// read once and reused for layout, the published step count, and the breadcrumb's
		// own K. ok is false exactly when the two centers coincide (EdgeCenterDistAndDir's
		// own guard) — skip this edge exactly as before.
		dist, liveDir, count, ok := beadindex.ChainEdgeGeometry(selfCenter, targetCenter, selfTorusR, m.geom.Kind, m.topo.neighborKinds[to])
		if !ok {
			continue
		}

		// Publish this edge's freshly computed step count onto its own *wire.Out
		// (docs/bead-model/bead-lattice.md "Ownership") and onto its edgeMover's stepsIn (so a live
		// in-flight bead's remaining travel — edgeMover.recomputeGeometry's
		// ReviseInFlightGeometry call — is revised against the same integer too; see
		// edge_mover.go's stepsIn doc comment for why a second delivery is needed instead of
		// the edgeMover reading the Out directly). Both are non-blocking, latest-wins sends —
		// this node's own goroutine never waits on either reader.
		for i, wt := range m.outs.outWireTargets {
			if wt != to {
				continue
			}
			if i < len(m.outs.outWireOuts) && m.outs.outWireOuts[i] != nil {
				m.outs.outWireOuts[i].PublishSteps(count)
			}
			if i < len(m.outs.outStepsIn) && m.outs.outStepsIn[i] != nil {
				stepdeliver.SendStepsNonBlocking(m.outs.outStepsIn[i], count)
			}
		}

		// Which bead this edge's traversals have reached. Read from THIS node's own
		// outgoing wire for this target, on this node's own goroutine (it is the goroutine
		// that drives that wire — see nodeMover.outWires), so LiveBeadFractions' single-
		// goroutine contract holds and no other goroutine's state is touched.
		//
		// index -> the traversing bead's VALUE. The value travels because the lit bead takes
		// bead 0's or bead 1's own fill: a bare "is lit" flag could not say which.
		// THE PULSES on this edge: one entry per live traversal, each carrying its own
		// CONTINUOUS fraction rather than the index of a chain bead it has reached.
		//
		// The edge is drawn as a line now, not as a sequence of placeholder beads lighting
		// in turn, so there is no slot for a traversal to occupy — it has a position. That
		// is why litBeadIndex (a FLOOR onto a bead index, lit_bead_index.go) is not used
		// here: floor is exactly what made the old motion a series of jumps, one bead-width
		// each, and no tick rate can smooth a value that is quantised before it is drawn.
		//
		// t stays the wire's own [0,1) fraction, and the step count still comes with the
		// bead (p.Steps), so lighting and layout still read ONE length — the property
		// litBeadIndex's doc comment protects, kept by passing t through instead of an index.
		// Gathered into beadindex.Pulse values (t, the step count that t was computed
		// against, and the carried value) — a plain data value from here on, handed to
		// ChainBeadRows below.
		var pulses []beadindex.Pulse
		for i, wt := range m.outs.outWireTargets {
			if wt != to || m.outs.outWires[i] == nil {
				continue
			}
			for _, p := range m.outs.outWires[i].LiveBeadFractions(tick) {
				if p.T < 0 || p.T >= 1 || p.Steps <= 0 {
					continue
				}
				// p.Steps travels WITH the bead: it is the length its own t was computed
				// against. Carried here so placement spans that same integer — see the
				// placement below.
				pulses = append(pulses, beadindex.Pulse{T: p.T, Steps: p.Steps, Val: int32(p.Val)})
			}
		}
		// DIAGNOSTIC ONLY (task/log-node4-chain-aim): one breadcrumb per outgoing target
		// per chainBeads() call. Gated on m.tr != nil exactly like emitGeometry's own
		// breadcrumb calls elsewhere in this package — cheap no-op with no stream wired
		// (headless tests, bare movers).
		// DIAGNOSTIC, AND OFF BY DEFAULT. chainBeads runs once per cycle per outgoing
		// target, so this breadcrumb is per-tick — 56,492 rows and 15 MB from a single
		// two-node run, which buries every control-event breadcrumb emitted alongside it
		// and is the exact firehose .claude/rules/go-debugging.md warns against ("keep it
		// SPARSE — it is a debug tool for control events, not a per-tick firehose").
		//
		// Gated at the SOURCE on one env var read once at process start, the same shape
		// WIREFOLD_EDGE_BEAD_TRACE uses for the other high-volume trace (paced_wire.go):
		// with it unset nothing is formatted and nothing is appended, rather than being
		// built and discarded downstream. Set WIREFOLD_CHAIN_AIM_TRACE=1 to get it back.
		// beadindex.ChainAimBreadcrumbText builds the pure VALUE string (a phase of its
		// own); the m.tr.Breadcrumb call right after it is the side effect — a send on the
		// trace channel — that formatting is not.
		if m.tr != nil && chainAimTraceEnabled {
			targetRow := int32(-1)
			if m.topo.nodeRowFor != nil {
				if r, ok := m.topo.nodeRowFor(to); ok {
					targetRow = r
				}
			}
			value := beadindex.ChainAimBreadcrumbText(to, count, dist, liveDir)
			m.tr.Breadcrumb("chain-aim", m.id, to, value)
			breadcrumbs = append(breadcrumbs, wire.RowEvent{
				Kind: T.KindBreadcrumb, Label: T.BreadcrumbChainAim, Debug: 1,
				NodeRow: m.stream.nodeRow, PortRow: -1, TargetRow: targetRow, TargetPortRow: -1,
				EdgeRow: -1, Slot: -1, Text: value,
			})
		}
		// One coordinate: bead index i. Offset from this node's centre is
		// selfTorusR + lattice.BeadTorusOuterR + i*lattice.BeadStepR (docs/bead-model/bead-lattice.md
		// "Placement"). "Beads never inside a node" falls out of this tangency, with no
		// clamp.
		step := lattice.BeadStepR
		base := selfTorusR + lattice.BeadTorusOuterR
		offsetAt := func(i int) float64 {
			return beadindex.BeadPlacementOffset(base, step, i)
		}
		// aimUnit is the live direction, carried as a plain unit vector: this is what
		// gets broadcast to this edge's bead-actor chain (reconcileBeadChain,
		// bead_chain.go), which resolves each bead's own position from it directly (one
		// hop, dependency depth 1 — no neighbour read). Bead 0's resolved position IS
		// this node's own "node -> first bead" polar vector (MODEL.md): owned by that
		// bead's own goroutine, never a second stored copy here.
		aimUnit := liveDir
		// A MUTUAL pair (this target also aims back here) offsets its own chain
		// perpendicular to the shared centre line so the two chains do not draw on top of
		// each other — see nodegeom.ParallelChainOffset (nodegeom/port_geometry.go) for why the direction is
		// measured in canonical id order and why each end can decide alone. Zero for every
		// ordinary edge, so a one-way chain is untouched. The vector math lives in
		// port_geometry.go because this file is guarded against it
		// (tools/network/beads/check-no-sqrt-in-chain-beads.sh), the same split nodegeom.EdgeCenterDistAndDir uses.
		var chainSep vec3
		if m.topo.mutualTargets[to] {
			if off, ok := nodegeom.ParallelChainOffset(m.id, to, selfCenter, targetCenter, m.geom.SceneCenter); ok {
				chainSep = off
			}
		}
		// Production call site for the bead-actor primitive (nodes/wire/bead_actor.go,
		// bead_wake_group.go): nil in every bare-literal test nodeMover, so this stays a
		// no-op there and chainBeads keeps its pure, synchronous, deterministic contract
		// (see beadTickFn's own doc comment). In production this reconciles this edge's
		// live *wire.Bead goroutine count to `count` and broadcasts fresh geometry when
		// the aim or count changed.
		var actorChain *edgeBeadChain
		if m.beads.beadTickFn != nil {
			actorChain = m.reconcileBeadChain(to, count, offsetAt, aimUnit)
		}
		// resolved/resolvedValid are the bead-actor chain's own already-drained snapshot
		// positions, parallel to placeholder index i — nil/empty when there is no live
		// bead-actor chain (beadTickFn nil, every bare-literal test nodeMover), in which
		// case beadindex.ChainBeadRows falls back to the formula for every index. This is
		// the only piece of actor state fed into ChainBeadRows; it is read here, once, on
		// this node's own goroutine, and handed down as a plain slice.
		var resolved []vec3
		var resolvedValid []bool
		if actorChain != nil {
			resolved = make([]vec3, len(actorChain.last))
			resolvedValid = actorChain.valid
			for i, s := range actorChain.last {
				resolved[i] = s.Position
			}
		}
		// ChainBeadRows is chainBeads' own two placement loops — the placeholder loop
		// (every row UNLIT; it is no longer what a traversal looks like, but it stays on
		// the wire because it is still the chain's own geometry, still what the bead-actor
		// goroutines resolve, and still what every placement rule in this file's tests is
		// written against — the renderer simply draws lit rows only) and the lit-pulse loop
		// (one LIT row per live pulse, appended after this edge's chain, at a CONTINUOUS
		// index rather than an integer one — see PulsePlacementOffset's own doc comment) —
		// unchanged arithmetic, named as a phase (beadindex/chain_edge_layout.go).
		edgeOX, edgeOY, edgeOZ, edgeLit, edgeLitVal := beadindex.ChainBeadRows(liveDir, chainSep, base, step, count, resolved, resolvedValid, pulses)
		ox = append(ox, edgeOX...)
		oy = append(oy, edgeOY...)
		oz = append(oz, edgeOZ...)
		lit = append(lit, edgeLit...)
		litVal = append(litVal, edgeLitVal...)
	}
	return ox, oy, oz, lit, litVal, breadcrumbs
}
