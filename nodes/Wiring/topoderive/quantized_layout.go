// quantized_layout.go — the quantized-layout derive phase, lifted out of
// nodes/Wiring/loader_layout.go (movedispatch-decomposition.md item 8's class: a pure
// derive phase that touches no md./buildCtx/actor type).
package topoderive

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/quantoffset"
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// ComputeQuantizedLayout resolves every node's quantoffset.QuantizedOffset — the stored
// quantITheta/quantIPhi/quantIR when ALL THREE are present (a scene saved under this
// model), otherwise the offset MEASURED from the node's current (pre-quantized,
// scenePolar-derived) center (an old scene, or a node whose scenePolar was
// hand-authored) — then recomputes every node's world center directly about the scene
// center (every node independent — no reference/parent) and overwrites nodeGeoms/centers
// with the result. Every later phase (reach radii, per-edge arc/segment, the movers
// seeded in buildMoveDispatch) therefore operates on the composed centers, and
// md.lq.quantizedLayout defaults to true (buildMoveDispatch) so the live drag path
// (RootMove) treats this same offset model as authoritative too.
// ComputeQuantizedLayout mutates centers and nodeGeoms IN PLACE (both are maps, so the
// caller's copies are updated directly — there is nothing to additionally return for
// those two) and returns the per-node quantized offsets it computed.
func ComputeQuantizedLayout(spec loadspec.TopoSpec, sphere geom.SceneSphere, centers map[string]wire.Vec3, nodeGeoms map[string]nodegeom.NodeGeom) map[string]quantoffset.QuantizedOffset {
	ids := make(map[string]bool, len(spec.Nodes))
	for _, n := range spec.Nodes {
		ids[n.ID] = true
	}

	// The scalar triple is the STORED quantI* when a scene was saved under this model
	// (all three present); otherwise it is MEASURED from the node's currently-loaded
	// (pre-quantized, scenePolar-derived) center — the fallback for an un-migrated node.
	// prior carries each node's stored per-node step constants (when present in the
	// spec) so measureScalars preserves them into the fallback-measured offset instead
	// of defaulting to global constants for a node that DOES have its own.
	prior := make(map[string]quantoffset.QuantizedOffset, len(spec.Nodes))
	for _, n := range spec.Nodes {
		o := quantoffset.QuantizedOffset{}
		if n.StepTheta != nil {
			o.CTheta = *n.StepTheta
		}
		if n.StepPhi != nil {
			o.CPhi = *n.StepPhi
		}
		if n.StepR != nil {
			o.CR = *n.StepR
		}
		prior[n.ID] = o
	}

	measured := quantoffset.MeasureScalars(centers, ids, sphere.Center, prior)
	offsets := make(map[string]quantoffset.QuantizedOffset, len(spec.Nodes))
	// exact marks nodes whose EXACT position was persisted as scenePolar (r,θ,φ). For
	// those, the loaded center (toNodeGeom placed it at exactly that polar) is the
	// authoritative position — it is NOT overwritten by the quantized reconstruction
	// below, so a dragged node reloads at exactly where it was dropped. The quantized
	// triple for such a node is just measured bookkeeping.
	exact := make(map[string]bool, len(spec.Nodes))
	for _, n := range spec.Nodes {
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
			o := quantoffset.QuantizedOffset{
				ITheta: *n.QuantITheta,
				IPhi:   *n.QuantIPhi,
				IR:     *n.QuantIR,
			}
			if n.StepTheta != nil {
				o.CTheta = *n.StepTheta
			}
			if n.StepPhi != nil {
				o.CPhi = *n.StepPhi
			}
			if n.StepR != nil {
				o.CR = *n.StepR
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
	// (quantoffset's stepTheta/stepPhi/stepR) before anything downstream reads
	// iTheta/iPhi/iR — a single choke point covering every branch above (stored triple,
	// measured fallback, prior-only fallback), rather than three separate call sites that
	// could drift. See NormalizeOffset's doc comment for why this must run at every load,
	// not just as a one-time file migration.
	for id, o := range offsets {
		offsets[id] = quantoffset.NormalizeOffset(o)
	}

	// Reconstruct world centers from the quantized triple ONLY for nodes without an exact
	// stored scenePolar (legacy / un-migrated). Nodes with an exact scenePolar keep the
	// verbatim loaded center — their drag position round-trips losslessly.
	derived := quantoffset.DeriveCenters(offsets, sphere.Center)
	for id, pos := range derived {
		if exact[id] {
			continue
		}
		centers[id] = pos
		if g, ok := nodeGeoms[id]; ok {
			nodegeom.SetNodeWorld(&g, pos)
			nodeGeoms[id] = g
		}
	}
	return offsets
}
