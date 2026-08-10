package Wiring

import (
	"testing"

	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

// TestPulseSpeedParity is a drift guard, not a restructure: lattice.PulseSpeedWuPerMs
// (nodes/wire/lattice/bead_lattice.go) and nodegeom.CurveParamPulseSpeedWuPerMs (curve_params.go,
// above) are two literal copies of the same constant. They can't share one
// definition because gen-node-defs' parseCurveParams reads the literal directly
// out of curve_params.go's source text to emit
// tools/topology-vscode/src/schema/curve-params.ts, and nodes/wire/lattice cannot import
// nodes/Wiring (that would be the leaf importing its own consumer). This test is
// the single-source-of-truth substitute: if either copy drifts, it fails loudly
// instead of silently producing two different pulse speeds.
func TestPulseSpeedParity(t *testing.T) {
	if lattice.PulseSpeedWuPerMs != nodegeom.CurveParamPulseSpeedWuPerMs {
		t.Fatalf("lattice.PulseSpeedWuPerMs=%v != nodegeom.CurveParamPulseSpeedWuPerMs=%v — the two literal copies have drifted",
			lattice.PulseSpeedWuPerMs, nodegeom.CurveParamPulseSpeedWuPerMs)
	}
}
