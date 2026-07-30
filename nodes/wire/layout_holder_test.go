package wire

import "testing"

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
