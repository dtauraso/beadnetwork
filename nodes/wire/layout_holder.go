// layout_holder.go — the per-node LOCAL POLAR list.
//
// Every domain double-link (a bidirectional edge pair A↔B) gives each endpoint
// its own LOCAL POLAR to the other, measured with ITSELF as center, in the same
// (quantITheta,quantIPhi,quantIR)×(stepTheta,stepPhi,stepR) integer-scalar form
// used for the node's absolute scene-polar position (quantized_layout.go). A node
// with N connections holds N local polars — one per neighbor, owned by that node
// alone (A's entry for B and B's entry for A are separate values).
//
// LayoutHolder is embedded into every node kind's struct (directly, or via
// gatecommon.GateNode for the two gate kinds) so loader.go can locate it by
// reflection (the same field-lookup used for port/data injection) and load the
// computed list through LoadLocalPolars.
package wire

import (
	"fmt"
	"math"
)

// Default local-polar quantization cells — SMALL and uniform across every node,
// unlike the scene-center triple's stepR=20/π-12 steps. The point of the
// double-link local-polar model is that a moved node's distance to each
// neighbor always lands on a WHOLE tick of that neighbor's own small grid;
// coarse scene-sized cells would leave most drags at iR=0/1 with no
// resolution. Used as the fallback whenever a LocalPolar has no stored
// per-neighbor step constants (LocalPolar.StepTheta/StepPhi/StepR == 0).
const (
	localStepTheta = math.Pi / 180 // 1 degree
	localStepPhi   = math.Pi / 180 // 1 degree
	// LocalStepR is exported (unlike its theta/phi siblings) because it is now
	// the BEAD lattice's own step — a node moves exactly ONE BEAD DISTANCE per
	// lattice tick, not a fraction of one. This used to be a SEPARATE, finer
	// lattice (2.0, hand-picked, with BeadStepCells=4 bead-lattice cells
	// nested inside each bead step) so a drag had sub-bead resolution; that
	// two-lattice design is what caused the placement bug this collapse fixes
	// (bead_lattice.go's BeadStepCells doc comment) — PLACEMENT read each
	// entry's own stored per-entry stepR (which drifted to 2.0, a leftover of
	// the old finer lattice) while the COUNT divided by the definition of
	// BeadStepCells (4) against the now-different LocalStepR (2.24), so the
	// two disagreed by 12% and surplus beads ran into the target node. With
	// one lattice there is nothing left to nest: LocalStepR IS BeadStepR, so
	// a node-lattice cell and a bead step are the same distance by
	// construction, not by two constants staying in sync. Every stored
	// quantIR keeps its cell-COUNT meaning, but each cell now measures the
	// bead lattice's own 8.96 (not the old finer 2.24), so the whole graph
	// expands versus what was on screen under the disagreeing pair — see
	// docs/bead-lattice.md for why that expansion is accepted, not
	// compensated for.
	LocalStepR = BeadStepR // world units; = 8.96 — one bead distance
	// localStepR is kept as a same-package alias so every other call site in
	// this file (localPolarSteps/EffectiveSteps, both pre-dating the bead
	// lattice) does not need touching just to read the new exported name.
	localStepR = LocalStepR
)

// DefaultLocalStepTheta/Phi/R are the exported mirrors of localStepTheta/Phi/R, for
// callers in another package (drag_persist_e2e_test.go in nodes/Wiring) that need to
// assert against the exact default local-polar grid constants.
const (
	DefaultLocalStepTheta = localStepTheta
	DefaultLocalStepPhi   = localStepPhi
	DefaultLocalStepR     = localStepR
)

// LocalPolar is one node's local-polar offset to a neighbor it shares a domain
// edge with, measured with the OWNING node as center, in the same integer-scalar
// form as quantizedOffset (quantized_layout.go).
type LocalPolar struct {
	To string // neighbor node id

	QuantITheta int
	QuantIPhi   int
	QuantIR     int

	// Per-neighbor step constants — same "own constants, default-fallback"
	// contract as quantizedOffset.cTheta/cPhi/cR. Zero means unset (falls back
	// to the package's local-polar defaults: localStepTheta/localStepPhi/localStepR).
	StepTheta float64
	StepPhi   float64
	StepR     float64
}

// effectiveSteps mirrors quantizedOffset.effectiveSteps: this local polar's own
// step constants, falling back to the SMALL local-polar defaults (NOT the scene
// triple's coarser stepTheta/stepPhi/stepR) for any unset component.
func (lp LocalPolar) EffectiveSteps() (t, p, r float64) {
	t, p, r = lp.StepTheta, lp.StepPhi, lp.StepR
	if t == 0 {
		t = localStepTheta
	}
	if p == 0 {
		p = localStepPhi
	}
	if r == 0 {
		r = localStepR
	}
	return
}

// LayoutHolder is embedded into every node kind's struct. It owns this node's
// LocalPolars list (one per domain-edge neighbor).
//
// Invariant: a LayoutHolder is written and read ONLY by its owning node's own
// goroutine. RootMove (node_move.go) routes a drag's moveMsgKindDrag to the
// DRAGGED NODE'S OWN inbox, so commitLocal -> requantizeLocalPolars runs on
// that node's own goroutine (nodeMover.handle). A node never reaches into a
// NEIGHBOR's LayoutHolder directly: it sends a moveMsgKindNeighborSetC message
// (via its own retry queue, onto the neighbor's directed neighborIn channel),
// and it is the neighbor's own run/handle goroutine that drains that message
// and calls neighborSetCRequantize -> lh.SetLocalPolar/SetPole on ITS OWN
// holder. One holder, one owning goroutine, neighbors reached only by
// message — no cross-goroutine access to guard against.
type LayoutHolder struct {
	localPolars []LocalPolar
	// pole is the measurement pole (rotating_pole.go localPole result) that
	// localPolars' current QuantITheta/QuantIPhi entries were last quantized about.
	// Persisted (WriteLocalPolars) so a reload reconstructs identical world directions
	// without re-deriving from live cartesian — see requantizePoleTraced's doc comment
	// in node_move.go for why this must be carried rather than recomputed from scratch
	// against an assumed home pole.
	pole Pole
}

// Pole is a direction on the unit sphere (pole = +y): θ = angle from +y (0=up,
// π=down), φ = azimuth around +y (0=+x, increasing toward +z). Structurally
// identical to nodes/Wiring's spherical.go `dir` (same shape, deliberately
// duplicated rather than shared — wire must not import Wiring, since Wiring
// already imports wire). Wiring converts at the LayoutHolder.Pole()/SetPole()
// call sites via a direct struct conversion (dir(lh.Pole()) / wire.Pole(d)),
// valid because the two types have identical underlying structure.
type Pole struct {
	Theta float64
	Phi   float64
}

// localPolarSteps returns the effective step constants of this node's CURRENT
// stored local polar to the given neighbor (falling back to the local-polar
// defaults if no entry exists yet), so a re-quantize preserves a neighbor's
// own step constants across drags exactly like quantizedOffset does for the
// scene triple.
func (lh *LayoutHolder) localPolarSteps(to string) (t, p, r float64) {
	for _, lp := range lh.localPolars {
		if lp.To == to {
			return lp.EffectiveSteps()
		}
	}
	return LocalPolar{}.EffectiveSteps()
}

// LocalPolarSteps is localPolarSteps' exported entry point for callers in another
// package (quantized_move.go's requantizePoleTraced in nodes/Wiring).
func (lh *LayoutHolder) LocalPolarSteps(to string) (t, p, r float64) {
	return lh.localPolarSteps(to)
}

// LoadLocalPolars replaces this node's entire local-polar list. Used
// exactly once, at load time (loader.go), to attach the freshly-computed list
// (computeLocalPolars) to the node's own LayoutHolder — the only initial-load
// writer, distinct from SetLocalPolar's per-neighbor upsert used by drags.
//
// QuantIR is stored verbatim — there is only one lattice now (SnapQuantIR no
// longer exists, bead_lattice.go), so a loaded separation has nothing to snap
// to but itself.
//
// Fails LOUDLY on a stored StepR that disagrees with LocalStepR, the one THIS
// bug actually was: an on-disk entry (topology/nodes/<id>/local-polars.json)
// used to carry its own "stepR" and PLACEMENT trusted it verbatim
// (LocalPolar.EffectiveSteps), while the edge-length COUNT was computed
// against a different, hardcoded assumption of what the lattice step was — so
// a stale stored constant silently overrode the lattice with nothing to
// notice the two had drifted apart. A LocalPolar with StepR unset (0) still
// falls back to LocalStepR via EffectiveSteps and is fine; a LocalPolar whose
// StepR is EXPLICITLY set to something else is exactly the shape of the bug
// that shipped, so it panics here instead of loading quietly.
// TestLoadLocalPolarsRejectsDisagreeingStepR pins this.
func (lh *LayoutHolder) LoadLocalPolars(lps []LocalPolar) {
	for _, lp := range lps {
		if lp.StepR != 0 && math.Abs(lp.StepR-LocalStepR) > 1e-9 {
			panic(fmt.Sprintf(
				"LoadLocalPolars: local polar to %q stored stepR=%v, want wire.LocalStepR=%v — "+
					"a per-entry step that disagrees with the lattice is the bead-penetration bug "+
					"this rejects (docs/bead-lattice.md); migrate the stored value instead of loading it",
				lp.To, lp.StepR, LocalStepR))
		}
	}
	lh.localPolars = lps
}

// SetLocalPolar upserts this node's local-polar entry for the given neighbor
// (updating in place if present, appending otherwise). The sole in-memory
// writer of LocalPolars outside load-time construction.
//
// quantIR is stored verbatim — there is only one lattice now (SnapQuantIR no
// longer exists, bead_lattice.go), so a live re-quantize (requantizePoleTraced,
// neighborSetCRequantize) already measures directly in bead-step-sized cells
// and has nothing left to snap to.
func (lh *LayoutHolder) SetLocalPolar(to string, quantITheta, quantIPhi, quantIR int, stepTheta, stepPhi, stepR float64) {
	for i := range lh.localPolars {
		if lh.localPolars[i].To == to {
			lh.localPolars[i].QuantITheta = quantITheta
			lh.localPolars[i].QuantIPhi = quantIPhi
			lh.localPolars[i].QuantIR = quantIR
			lh.localPolars[i].StepTheta = stepTheta
			lh.localPolars[i].StepPhi = stepPhi
			lh.localPolars[i].StepR = stepR
			return
		}
	}
	lh.localPolars = append(lh.localPolars, LocalPolar{
		To: to, QuantITheta: quantITheta, QuantIPhi: quantIPhi, QuantIR: quantIR,
		StepTheta: stepTheta, StepPhi: stepPhi, StepR: stepR,
	})
}

// LocalPolarsSnapshot returns a defensive copy of this node's current
// LocalPolars list, safe to hand to a persister running on another goroutine.
func (lh *LayoutHolder) LocalPolarsSnapshot() []LocalPolar {
	out := make([]LocalPolar, len(lh.localPolars))
	copy(out, lh.localPolars)
	return out
}

// Pole returns the measurement pole this node's current LocalPolars entries were last
// quantized about (world +y, dir{0,0}, if never set — the home pole default).
func (lh *LayoutHolder) Pole() Pole {
	return lh.pole
}

// SetPole records the measurement pole the CURRENT LocalPolars entries were quantized
// about, so a later requantize (or a reload) can reconstruct an unchanged neighbor's
// direction from its stored indices without re-measuring live cartesian geometry.
func (lh *LayoutHolder) SetPole(p Pole) {
	lh.pole = p
}
