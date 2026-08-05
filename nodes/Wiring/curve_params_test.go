package Wiring

import (
	"math"
	"testing"
)

// TestVectorAngleStepIsPiOver12 pins CurveParamVectorAngleStep's bare float literal
// (required by gen-node-defs' CurveParam* extractor — see the constant's own doc
// comment) against math.Pi/12, so the two can never silently diverge.
func TestVectorAngleStepIsPiOver12(t *testing.T) {
	want := math.Pi / 12
	if math.Abs(CurveParamVectorAngleStep-want) > 1e-15 {
		t.Fatalf("CurveParamVectorAngleStep = %v, want math.Pi/12 = %v", CurveParamVectorAngleStep, want)
	}
}
