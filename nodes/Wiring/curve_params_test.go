package Wiring

import (
	"math"
	"testing"
)

// TestTiltVectorAngleStepIsPiOver12 pins CurveParamTiltVectorAngleStep's bare float
// literal (required by gen-node-defs' CurveParam* extractor — see the constant's own doc
// comment) against math.Pi/12, so the two can never silently diverge.
func TestTiltVectorAngleStepIsPiOver12(t *testing.T) {
	want := math.Pi / 12
	if math.Abs(CurveParamTiltVectorAngleStep-want) > 1e-15 {
		t.Fatalf("CurveParamTiltVectorAngleStep = %v, want math.Pi/12 = %v", CurveParamTiltVectorAngleStep, want)
	}
}
