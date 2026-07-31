package wire

import (
	"math"
	"testing"
)

// TestLoadLocalPolarsNormalizesDisagreeingStepR pins what LoadLocalPolars does with a
// stored StepR that disagrees with LocalStepR: it OVERWRITES it with the lattice's own
// value rather than loading it. That silent win — an on-disk "stepR": 2, left over from
// the earlier finer lattice, quietly overriding LocalStepR (since collapsed onto the bead
// lattice's 8.96) — is the exact defect docs/bead-lattice.md diagnoses: placement trusted
// the stale stored constant while the edge-length count assumed the current one, so chain
// beads over-budgeted and ran into the target node.
//
// This check REJECTED the disagreement by panicking first, which was wrong for a reason
// worth keeping: the editor's runner respawns the process on exit, so a panic at load is
// not a loud failure but a crash loop — it reached the live editor as an unreadable
// flicker with no scene at all. Normalizing is both safe and better: the lattice is the
// authority, so there is nothing for a human to arbitrate.
func TestLoadLocalPolarsNormalizesDisagreeingStepR(t *testing.T) {
	lh := &LayoutHolder{}
	lh.LoadLocalPolars([]LocalPolar{
		{To: "neighbor", QuantIR: 10, StepR: 2}, // stale, pre-collapse lattice's step
	})
	got := lh.LocalPolarsSnapshot()
	if len(got) != 1 {
		t.Fatalf("want 1 loaded entry, got %d", len(got))
	}
	if got[0].StepR != LocalStepR {
		t.Fatalf("stale stepR=2 loaded as %v, want it normalized to LocalStepR=%v — a per-entry step that disagrees with the lattice is the bead-penetration bug",
			got[0].StepR, LocalStepR)
	}
	// The DISTANCE the entry meant must survive. QuantIR and StepR are one value
	// (QuantIR*StepR); converting the step alone multiplied every separation by
	// LocalStepR/StepR — 4.5x for a stale 2.0 — which is how a 25-bead edge got 128.
	const wantIR = 2 // round(10 * 2 / 8.96)
	if got[0].QuantIR != wantIR {
		t.Fatalf("QuantIR=%d after normalizing stepR 2 -> %v, want %d: the stored distance was %v world units and is now %v — normalizing half the pair changes the geometry",
			got[0].QuantIR, LocalStepR, wantIR, 10*2.0, float64(got[0].QuantIR)*LocalStepR)
	}
}

// TestLoadLocalPolarsLeavesAgreeingOrUnsetStepR is the companion negative case: a stored
// StepR that already matches LocalStepR, or is simply unset (0, falling back to LocalStepR
// via EffectiveSteps), must pass through untouched. Normalization must correct
// DISAGREEMENT, not rewrite every entry — an unset value carries the "no opinion, use the
// default" meaning EffectiveSteps depends on, and filling it in would erase that.
func TestLoadLocalPolarsLeavesAgreeingOrUnsetStepR(t *testing.T) {
	lh := &LayoutHolder{}
	lh.LoadLocalPolars([]LocalPolar{
		{To: "a", QuantIR: 10, StepR: LocalStepR},
		{To: "b", QuantIR: 10}, // StepR unset
	})
	got := lh.LocalPolarsSnapshot()
	if len(got) != 2 {
		t.Fatalf("want 2 loaded entries, got %d", len(got))
	}
	if got[0].StepR != LocalStepR {
		t.Fatalf("agreeing stepR changed to %v, want %v", got[0].StepR, LocalStepR)
	}
	if got[1].StepR != 0 {
		t.Fatalf("unset stepR filled in as %v, want it left at 0 so EffectiveSteps still supplies the default", got[1].StepR)
	}
}

// TestLoadLocalPolarsNormalizesDisagreeingAngularSteps is StepR's companion for the two
// angular axes: a stale StepTheta/StepPhi that disagrees with the current
// localStepTheta/localStepPhi default must be normalized the same way — index and step
// rescaled TOGETHER so the represented ANGLE survives, not just the step constant
// overwritten in place (which would silently rescale every stored bearing).
func TestLoadLocalPolarsNormalizesDisagreeingAngularSteps(t *testing.T) {
	lh := &LayoutHolder{}
	const staleTheta = 0.3 // stale, coarser than the current 1-degree default
	const stalePhi = 0.34  // observed on disk: nodes 3/5/8's ~19.5deg stepPhi
	lh.LoadLocalPolars([]LocalPolar{
		{To: "neighbor", QuantITheta: 7, StepTheta: staleTheta, QuantIPhi: 4, StepPhi: stalePhi},
	})
	got := lh.LocalPolarsSnapshot()
	if len(got) != 1 {
		t.Fatalf("want 1 loaded entry, got %d", len(got))
	}
	if got[0].StepTheta != localStepTheta {
		t.Fatalf("stale stepTheta=%v loaded as %v, want normalized to localStepTheta=%v", staleTheta, got[0].StepTheta, localStepTheta)
	}
	if got[0].StepPhi != localStepPhi {
		t.Fatalf("stale stepPhi=%v loaded as %v, want normalized to localStepPhi=%v", stalePhi, got[0].StepPhi, localStepPhi)
	}
	wantITheta := int(math.Round(7 * staleTheta / localStepTheta))
	if got[0].QuantITheta != wantITheta {
		t.Fatalf("QuantITheta=%d after normalizing stepTheta %v -> %v, want %d: the stored bearing was %v rad and is now %v rad — rewriting only the step changes the represented angle",
			got[0].QuantITheta, staleTheta, localStepTheta, wantITheta, 7*staleTheta, float64(got[0].QuantITheta)*localStepTheta)
	}
	wantIPhi := int(math.Round(4 * stalePhi / localStepPhi))
	if got[0].QuantIPhi != wantIPhi {
		t.Fatalf("QuantIPhi=%d after normalizing stepPhi %v -> %v, want %d", got[0].QuantIPhi, stalePhi, localStepPhi, wantIPhi)
	}
}

// TestLoadLocalPolarsLeavesAgreeingOrUnsetAngularSteps is the angular-axis negative case,
// mirroring TestLoadLocalPolarsLeavesAgreeingOrUnsetStepR: an already-agreeing or unset (0)
// StepTheta/StepPhi must pass through untouched, since 0 carries the "no opinion, use the
// default" meaning EffectiveSteps depends on.
func TestLoadLocalPolarsLeavesAgreeingOrUnsetAngularSteps(t *testing.T) {
	lh := &LayoutHolder{}
	lh.LoadLocalPolars([]LocalPolar{
		{To: "a", QuantITheta: 5, StepTheta: localStepTheta, QuantIPhi: 5, StepPhi: localStepPhi},
		{To: "b", QuantITheta: 5, QuantIPhi: 5}, // StepTheta/StepPhi unset
	})
	got := lh.LocalPolarsSnapshot()
	if len(got) != 2 {
		t.Fatalf("want 2 loaded entries, got %d", len(got))
	}
	if got[0].StepTheta != localStepTheta || got[0].StepPhi != localStepPhi {
		t.Fatalf("agreeing angular steps changed: theta=%v phi=%v, want %v/%v", got[0].StepTheta, got[0].StepPhi, localStepTheta, localStepPhi)
	}
	if got[0].QuantITheta != 5 || got[0].QuantIPhi != 5 {
		t.Fatalf("agreeing entry's indices changed: theta=%d phi=%d, want 5/5", got[0].QuantITheta, got[0].QuantIPhi)
	}
	if got[1].StepTheta != 0 || got[1].StepPhi != 0 {
		t.Fatalf("unset angular steps filled in: theta=%v phi=%v, want left at 0 so EffectiveSteps still supplies the default", got[1].StepTheta, got[1].StepPhi)
	}
}

// TestLoadLocalPolarsRealDivergenceStepTheta reproduces the exact live-data divergence
// found in topology/nodes/4/local-polars.json's entry for neighbor 6: stepTheta 0.11111
// (6.37deg, stale) with quantITheta 13 — while node 6's reciprocal entry for 4 already sits
// on the current default (0.017453, 1deg). LoadLocalPolars must rescale node 4's stale
// entry onto the SAME 0.017453 lattice with the bearing preserved, which is what lets both
// ends land on one shared step constant (see the two-end-agreement test below).
func TestLoadLocalPolarsRealDivergenceStepTheta(t *testing.T) {
	lh := &LayoutHolder{}
	const staleTheta = 0.11111
	const quantITheta = 13
	lh.LoadLocalPolars([]LocalPolar{
		{To: "6", QuantITheta: quantITheta, StepTheta: staleTheta},
	})
	got := lh.LocalPolarsSnapshot()
	if got[0].StepTheta != localStepTheta {
		t.Fatalf("node 4's live-data stepTheta=%v normalized to %v, want the current default %v", staleTheta, got[0].StepTheta, localStepTheta)
	}
	wantITheta := int(math.Round(quantITheta * staleTheta / localStepTheta))
	if got[0].QuantITheta != wantITheta {
		t.Fatalf("QuantITheta=%d after normalizing the live 4->6 stepTheta, want %d (bearing %v rad preserved, now %v rad)",
			got[0].QuantITheta, wantITheta, quantITheta*staleTheta, float64(got[0].QuantITheta)*localStepTheta)
	}
}

// TestLoadLocalPolarsBothEndsOfEdgeAgreeAfterNormalize is the invariant this whole change
// exists for and that had NO coverage before it: the same physical edge, loaded from two
// separate on-disk records that disagreed on their stored step (the live 4<->6 divergence —
// node 4's entry at 0.11111, node 6's reciprocal entry already at 0.017453), must end up on
// the SAME step constant after LoadLocalPolars runs on each node's own holder. Two
// independently-normalized entries landing on different constants would mean the two ends
// of one edge still disagree on what angle a tick represents — exactly the bug that let the
// chain snap between sides of the true line.
func TestLoadLocalPolarsBothEndsOfEdgeAgreeAfterNormalize(t *testing.T) {
	node4 := &LayoutHolder{}
	node4.LoadLocalPolars([]LocalPolar{
		{To: "6", QuantITheta: 13, StepTheta: 0.11111},
	})
	node6 := &LayoutHolder{}
	node6.LoadLocalPolars([]LocalPolar{
		{To: "4", QuantITheta: 9, StepTheta: 0.017453},
	})
	got4 := node4.LocalPolarsSnapshot()
	got6 := node6.LocalPolarsSnapshot()
	if got4[0].StepTheta != got6[0].StepTheta {
		t.Fatalf("both ends of edge 4<->6 must share one step constant after normalize: node4=%v node6=%v", got4[0].StepTheta, got6[0].StepTheta)
	}
	if got4[0].StepTheta != localStepTheta {
		t.Fatalf("normalized shared step=%v, want the current default localStepTheta=%v", got4[0].StepTheta, localStepTheta)
	}
}
