package Wiring

import (
	"testing"

	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// TestPulseSpeedParity is a drift guard, not a restructure: wire.PulseSpeedWuPerMs
// (nodes/wire/paced_wire.go) and CurveParamPulseSpeedWuPerMs (curve_params.go,
// above) are two literal copies of the same constant. They can't share one
// definition because gen-node-defs' parseCurveParams reads the literal directly
// out of curve_params.go's source text to emit
// tools/topology-vscode/src/schema/curve-params.ts, and nodes/wire cannot import
// nodes/Wiring (that would be the leaf importing its own consumer). This test is
// the single-source-of-truth substitute: if either copy drifts, it fails loudly
// instead of silently producing two different pulse speeds.
func TestPulseSpeedParity(t *testing.T) {
	if wire.PulseSpeedWuPerMs != CurveParamPulseSpeedWuPerMs {
		t.Fatalf("wire.PulseSpeedWuPerMs=%v != CurveParamPulseSpeedWuPerMs=%v — the two literal copies have drifted",
			wire.PulseSpeedWuPerMs, CurveParamPulseSpeedWuPerMs)
	}
}
