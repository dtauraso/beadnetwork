package Wiring

// node's quantizedOffset — the stored quantITheta/quantIPhi/quantIR when ALL THREE are
// present (a scene saved under this model), otherwise the offset MEASURED from the
// node's current (pre-quantized) scenePolar-derived center (an old scene, or a node
// whose scenePolar was hand-authored) — then recomputes every node's world center
// directly about the scene center (every node independent — no reference/parent) and
// overwrites b.nodeGeoms/b.centers with the result. Every later phase (reach radii,
// per-edge arc/segment, the movers seeded in buildMoveDispatch) therefore operates on
// the composed centers, and md.lq.quantizedLayout defaults to true (buildMoveDispatch) so
// the live drag path (RootMove) treats this same offset model as authoritative too.
func (b *buildCtx) computeQuantizedLayout() {
	ids := make(map[string]bool, len(b.spec.Nodes))
	for _, n := range b.spec.Nodes {
		ids[n.ID] = true
	}

	// The scalar triple is the STORED quantI* when a scene was saved under this model
	// (all three present); otherwise it is MEASURED from the node's currently-loaded
	// (pre-quantized, scenePolar-derived) center — the fallback for an un-migrated node.
	// prior carries each node's stored per-node step constants (when present in the
	// spec) so measureScalars preserves them into the fallback-measured offset instead
	// of defaulting to global constants for a node that DOES have its own.
	prior := make(map[string]quantizedOffset, len(b.spec.Nodes))
	for _, n := range b.spec.Nodes {
		o := quantizedOffset{}
		if n.StepTheta != nil {
			o.cTheta = *n.StepTheta
		}
		if n.StepPhi != nil {
			o.cPhi = *n.StepPhi
		}
		if n.StepR != nil {
			o.cR = *n.StepR
		}
		prior[n.ID] = o
	}

	measured := measureScalars(b.centers, ids, b.sphere.Center, prior)
	offsets := make(map[string]quantizedOffset, len(b.spec.Nodes))
	// exact marks nodes whose EXACT position was persisted as scenePolar (r,θ,φ). For
	// those, the loaded center (toNodeGeom placed it at exactly that polar) is the
	// authoritative position — it is NOT overwritten by the quantized reconstruction
	// below, so a dragged node reloads at exactly where it was dropped. The quantized
	// triple for such a node is just measured bookkeeping.
	exact := make(map[string]bool, len(b.spec.Nodes))
	for _, n := range b.spec.Nodes {
		if n.ScenePolarR != nil && n.ScenePolarTheta != nil && n.ScenePolarPhi != nil {
			exact[n.ID] = true
			if off, ok := measured[n.ID]; ok {
				offsets[n.ID] = off
			} else {
				offsets[n.ID] = prior[n.ID]
			}
			continue
		}
		if n.QuantITheta != nil && n.QuantIPhi != nil && n.QuantIR != nil {
			o := quantizedOffset{
				iTheta: *n.QuantITheta,
				iPhi:   *n.QuantIPhi,
				iR:     *n.QuantIR,
			}
			if n.StepTheta != nil {
				o.cTheta = *n.StepTheta
			}
			if n.StepPhi != nil {
				o.cPhi = *n.StepPhi
			}
			if n.StepR != nil {
				o.cR = *n.StepR
			}
			offsets[n.ID] = o
			continue
		}
		if off, ok := measured[n.ID]; ok {
			offsets[n.ID] = off
			continue
		}
		offsets[n.ID] = prior[n.ID] // centerless → default to the scene center, keep any stored constants
	}
	// NORMALIZE every offset's per-axis step constants against the CURRENT scene lattice
	// (quantized_layout.go stepTheta/stepPhi/stepR) before anything downstream reads
	// iTheta/iPhi/iR — a single choke point covering every branch above (stored triple,
	// measured fallback, prior-only fallback), rather than three separate call sites that
	// could drift. See normalizeOffset's doc comment for why this must run at every load,
	// not just as a one-time file migration.
	for id, o := range offsets {
		offsets[id] = normalizeOffset(o)
	}
	b.quantizedOffsets = offsets

	// Reconstruct world centers from the quantized triple ONLY for nodes without an exact
	// stored scenePolar (legacy / un-migrated). Nodes with an exact scenePolar keep the
	// verbatim loaded center — their drag position round-trips losslessly.
	derived := deriveCenters(offsets, b.sphere.Center)
	for id, pos := range derived {
		if exact[id] {
			continue
		}
		b.centers[id] = pos
		if g, ok := b.nodeGeoms[id]; ok {
			setNodeWorld(&g, pos)
			b.nodeGeoms[id] = g
		}
	}
}

// computeReachRadii computes each node's REACH radius (max distance from its
// center to any node it outputs to) under the loaded centers — non-rooted layout
// — streamed in NodeGeometry's sphereR field so the TS SphereRing reaches every
// surface node. Computed before newMoveDispatch so each node/edge mover captures
// it in its held geom.
func (b *buildCtx) computeReachRadii() {
	edges := make([]sphereEdge, 0, len(b.spec.Edges))
	for _, e := range b.spec.Edges {
		edges = append(edges, sphereEdge{Source: e.Source, Target: e.Target})
	}
	polars := map[string]polar{}
	for id, g := range b.nodeGeoms {
		if g.HasPos {
			polars[id] = g.ScenePolar
		}
	}
	for id, r := range reachRFromPolar(polars, edges) {
		g := b.nodeGeoms[id]
		g.ReachR = r
		b.nodeGeoms[id] = g
	}
}
