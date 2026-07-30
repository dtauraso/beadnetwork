package wire

import "testing"

// TestLoadLocalPolarsRejectsDisagreeingStepR pins the load-time check
// LoadLocalPolars' doc comment describes: a stored StepR that disagrees with
// LocalStepR must fail LOUDLY, not win silently. That silent win — an
// on-disk "stepR": 2 (left over from an earlier, finer lattice) quietly
// overriding LocalStepR (which had since become the bead lattice's own
// 8.96) — is the exact defect docs/bead-lattice.md's "THE BUG ON SCREEN"
// diagnoses: placement trusted the stale stored constant while the
// edge-length count assumed the current one, so chain beads over-budgeted
// and ran into the target node. This test proves the load path can no
// longer let that disagreement back in quietly.
func TestLoadLocalPolarsRejectsDisagreeingStepR(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("LoadLocalPolars: want a panic on a stored stepR that disagrees with LocalStepR, got none")
		}
	}()
	lh := &LayoutHolder{}
	lh.LoadLocalPolars([]LocalPolar{
		{To: "neighbor", QuantIR: 10, StepR: 2}, // stale, pre-collapse lattice's step
	})
}

// TestLoadLocalPolarsAcceptsAgreeingOrUnsetStepR is the companion positive case:
// a stored StepR that matches LocalStepR, or is simply unset (0, falling back to
// LocalStepR via EffectiveSteps), must load without panicking — the check must
// reject DISAGREEMENT, not every stored value.
func TestLoadLocalPolarsAcceptsAgreeingOrUnsetStepR(t *testing.T) {
	lh := &LayoutHolder{}
	lh.LoadLocalPolars([]LocalPolar{
		{To: "a", QuantIR: 10, StepR: LocalStepR},
		{To: "b", QuantIR: 10}, // StepR unset
	})
	if len(lh.LocalPolarsSnapshot()) != 2 {
		t.Fatalf("want 2 loaded entries, got %d", len(lh.LocalPolarsSnapshot()))
	}
}
