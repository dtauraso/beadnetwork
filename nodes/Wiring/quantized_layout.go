package Wiring

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"math"
)

// quantized_layout.go — the quantized FLAT ABSOLUTE SCENE-POLAR layout: every node is
// positioned independently by its own quantized polar offset — three integers
// (iTheta, iPhi, iR) — measured about the ONE scene-sphere center (no reference/parent
// concept; every node is a "root"). Three GLOBAL step constants (same for every node in
// the graph) turn the integers into a world offset.

// Default quantization step constants — used for any node that has no stored per-node
// step constants (quantizedOffset.cTheta/cPhi/cR == 0). Offsets are always integer
// multiples of a node's EFFECTIVE constants (its own, or these defaults):
// offset = (iTheta*cTheta, iPhi*cPhi, iR*cR).
//
// This used to be a SEPARATE, coarser lattice (15°/20.0) than the local-polar lattice a
// node's own chain beads are laid out from (layout_holder.go, 1°/8.96) — the finding in
// docs/which-lattice-a-node-lives-on.md: a node could not sit exactly on both, so its
// drawn position (this file) glided continuously while its beads jumped one bead distance
// at a time, and the two visibly slid against each other during a drag. Collapsing onto
// the SAME lattice (below) is the fix: a node's absolute position now moves by the same
// bead-distance/1-degree tick its neighbor distances already did.
const (
	// stepTheta/stepPhi match the local-polar lattice's own angular cell exactly
	// (layout_holder.go localStepTheta/localStepPhi) rather than an independently
	// hand-picked value (was π/12 = 15°) — so a node's absolute bearing and its
	// per-neighbor bearing tick together. An angle is NOT an exact bead distance at
	// every radius (arc = r*Δθ, so a fixed angle spans a different world distance at a
	// different r) — this is an accepted APPROXIMATION, not exact tangency the way
	// stepR is below: at the radii this graph occupies (~60-250 world units), 1° is
	// roughly 1-4.4 world units, the same order as a bead (8.96). A radius-dependent
	// step was rejected: it would make one stored angular index mean a different
	// angle at a different radius, which is worse than this approximation.
	stepTheta = math.Pi / 180
	stepPhi   = math.Pi / 180
	// stepR is now literally wire.BeadStepR (8.96) — one bead distance — not an
	// independently hand-picked value (was 20.0, chosen only to keep nodes distinct
	// at the ~80-unit spacing this graph happened to have). Derived, not copied as a
	// literal: bead_lattice.go owns the authored primitive (BeadRadius) this falls
	// out of, and this constant must move with it, never drift from a second copy.
	stepR = wire.BeadStepR
)

// quantizedOffset is a node's quantized polar offset (iTheta,iPhi,iR) about the ONE
// scene-sphere center, PLUS that node's own step constants (cTheta,cPhi,cR). iTheta/
// iPhi/iR default to zero (at the scene center) until authored. cTheta/cPhi/cR default
// to zero, meaning "unset" — effectiveSteps falls back to the global defaults
// (stepTheta/stepPhi/stepR) for any unset component, so an all-zero quantizedOffset
// reproduces today's global-constant behavior exactly.
type quantizedOffset struct {
	iTheta int
	iPhi   int
	iR     int

	cTheta float64
	cPhi   float64
	cR     float64
}

// effectiveSteps returns this node's own step constants, falling back to the global
// defaults for any component that is unset (zero). Every site that turns a scalar
// triple into (or out of) a world offset MUST go through this — it is the one place
// "per-node step, default fallback" is implemented.
func (o quantizedOffset) effectiveSteps() (t, p, r float64) {
	t, p, r = o.cTheta, o.cPhi, o.cR
	if t == 0 {
		t = stepTheta
	}
	if p == 0 {
		p = stepPhi
	}
	if r == 0 {
		r = stepR
	}
	return
}

// measureScalars is the flat-polar INVERSE measurement: given each node's current world
// center (centers), derive the integer scalar triple (iTheta, iPhi, iR) that is the
// node's polar coordinate about the ONE scene center — the model this file implements
// (see the package-level Model doc in node_move.go / CLAUDE.md). Every node is measured
// independently; there is no reference/parent origin.
//
// ids selects which node ids to measure (so callers can measure a subset without
// building a throwaway map); a node missing a center is omitted (nothing to measure).
//
// prior carries each node's PRIOR quantizedOffset (e.g. md.quantizedOffsets before a
// drag, or the loaded/measured offsets so far) so its stored step constants
// (cTheta/cPhi/cR) can be PRESERVED into the result — a node's constants never change
// on drag/remeasure, only its integer scalars do. prior may be nil (constants default
// to unset → global defaults).
func measureScalars(centers map[string]vec3, ids map[string]bool, sceneCenter vec3, prior map[string]quantizedOffset) map[string]quantizedOffset {
	result := make(map[string]quantizedOffset, len(ids))
	for id := range ids {
		pos, ok := centers[id]
		if !ok {
			continue
		}
		carried := prior[id] // zero value if absent — constants default to unset
		t, p_, r := carried.effectiveSteps()
		p := cart2polar(pos.Sub(sceneCenter))
		result[id] = quantizedOffset{
			iTheta: int(math.Round(p.Theta / t)),
			iPhi:   int(math.Round(p.Phi / p_)),
			iR:     int(math.Round(p.R / r)),
			cTheta: carried.cTheta,
			cPhi:   carried.cPhi,
			cR:     carried.cR,
		}
	}
	return result
}

// measureScalar is the single-node variant of measureScalars: given ONE node's ALREADY-
// COMPUTED polar position p (about the scene sphere center — cart2polar happens once, at
// the mouse-derived cartesian->polar boundary; see commitNodeMoveLocal), derive its
// integer scalar triple (iTheta, iPhi, iR), preserving prior's stored step constants
// (cTheta/cPhi/cR) exactly as measureScalars does. Used by the per-node commit path
// (commitNodeMoveLocal) so each node's quantized offset lives on that node's OWN mover
// (nodeMover.quantOffset) rather than a shared map read/written from multiple mover
// goroutines — see node6-drag-decentralized.md / the quantizedOffsets data-race fix.
func measureScalar(p polar, prior quantizedOffset) quantizedOffset {
	t, p_, r := prior.effectiveSteps()
	return quantizedOffset{
		iTheta: int(math.Round(p.Theta / t)),
		iPhi:   int(math.Round(p.Phi / p_)),
		iR:     int(math.Round(p.R / r)),
		cTheta: prior.cTheta,
		cPhi:   prior.cPhi,
		cR:     prior.cR,
	}
}

// offsetScenePolar is the index->position arithmetic half of the flat-polar FORWARD
// computation, factored out so a caller that needs the POLAR (not yet a cartesian world
// point) — commitNodeMoveLocal, to draw the quantized position it just measured — can
// reuse the exact same formula deriveCenters uses, rather than a second copy of
// (iR*cR, iTheta*cTheta, iPhi*cPhi) that could drift from this one:
//
//	offsetScenePolar(o) = {R: iR*cR, Theta: iTheta*cTheta, Phi: iPhi*cPhi}
//
// using o's OWN effective step constants (o.effectiveSteps()), falling back to the
// global defaults for any unset component.
func offsetScenePolar(o quantizedOffset) polar {
	t, p, r := o.effectiveSteps()
	return polar{R: float64(o.iR) * r, Theta: float64(o.iTheta) * t, Phi: float64(o.iPhi) * p}
}

// deriveCenters is the flat-polar FORWARD computation: given each node's scalar triple
// (scalars, from measureScalars or loaded meta.json quantI*), compute every node's world
// center directly about the ONE scene center — every node is independent (no reference/
// parent to resolve first): derived[id] = sceneCenter + polar2cart(offsetScenePolar(o)).
func deriveCenters(scalars map[string]quantizedOffset, sceneCenter vec3) map[string]vec3 {
	derived := make(map[string]vec3, len(scalars))
	for id, o := range scalars {
		derived[id] = sceneCenter.Add(polar2cart(offsetScenePolar(o)))
	}
	return derived
}

// normalizeOffset converts a quantized offset loaded with STALE per-axis step constants
// (from an older, coarser scene lattice — docs/which-lattice-a-node-lives-on.md) to the
// CURRENT step constants, preserving each axis's world distance/angle exactly the way
// LayoutHolder.LoadLocalPolars does for the local-polar lattice: new index = round(old
// index * old step / new step), and the step is rewritten ALONGSIDE the index, never left
// disagreeing with it (the two are one value — an index times its own step — and rewriting
// only one multiplies the represented distance by the wrong factor). A step of 0 already
// falls back to the current default via effectiveSteps and needs no conversion.
//
// This MUST run at every load, not just as a one-time file migration: each node's own
// mover rewrites its position.json from in-memory state on its next commit
// (persistQuantOffset), so a stale on-disk value a migration pass corrected would be put
// straight back by a running editor — the exact failure LoadLocalPolars' doc comment
// records happening to the local-polar migration this mirrors. Normalizing at load instead
// self-heals: the next persist writes the current, correct step forward.
func normalizeOffset(o quantizedOffset) quantizedOffset {
	if o.cTheta != 0 && math.Abs(o.cTheta-stepTheta) > 1e-9 {
		o.iTheta = int(math.Round(float64(o.iTheta) * o.cTheta / stepTheta))
		o.cTheta = stepTheta
	}
	if o.cPhi != 0 && math.Abs(o.cPhi-stepPhi) > 1e-9 {
		o.iPhi = int(math.Round(float64(o.iPhi) * o.cPhi / stepPhi))
		o.cPhi = stepPhi
	}
	if o.cR != 0 && math.Abs(o.cR-stepR) > 1e-9 {
		o.iR = int(math.Round(float64(o.iR) * o.cR / stepR))
		o.cR = stepR
	}
	return o
}
