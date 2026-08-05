// curve_params.go — single source of truth for curve-shape constants shared
// between the Go network and the TS visual layer.
//
// Codegen: tools/gen-node-defs reads this file and emits
// tools/topology-vscode/src/schema/curve-params.ts.
// After changing any constant here, regenerate with:
//   cd tools/topology-vscode && npm run gen:node-defs
//
// curve-params constants are prefixed with CurveParam so gen-node-defs can
// identify them via the "CurveParam" name prefix.

package Wiring

import (
	wire "github.com/dtauraso/wirefold/nodes/wire"
)

// CurveParamPulseSpeedWuPerMs is the uniform pulse speed in world-units per
// millisecond.  Both Go (simLatencyMs) and TS visual layer (travel
// duration) derive timing from this value.
//
// Matches PULSE_SPEED_WU_PER_MS in the generated curve-params.ts.
const CurveParamPulseSpeedWuPerMs = 0.04

// CurveParamNodeRadiusDivisor is the divisor applied to min(width,height)
// to obtain the node sphere radius.  Matches nodeRadius in geometry-helpers.ts
// (Math.min(width, height) / 4); port endpoints sit on this sphere surface.
const CurveParamNodeRadiusDivisor = 4

// CurveParamVectorAngleStep is the ONE quantization step for a node's own vector
// direction (Buffer/layout.go's VectorTheta/VectorPhi): every node's vectorThetaIdx/
// vectorPhiIdx (node_mover.go) is an INTEGER count of this step, never a free float
// (memory/feedback_abc_times_constant_not_rederive.md) — a single edit here changes the
// step for every node's vector in every scene. Same shape as the ring/pole angle
// vocabulary (polar.go): θ from world +y, φ azimuth around +y, radians. Also the display
// step the per-node vector-angle panel (NodeVectorAnglePanel.tsx) formats its
// θ/φ readout against — matches CURVE_PARAM_VECTOR_ANGLE_STEP in the generated
// curve-params.ts, so the panel's "5π/12"-style label can never drift from Go's own step.
//
// π/12 (15°) as a bare float literal, not the `math.Pi / 12` expression: gen-node-defs'
// CurveParam* extractor (tools/gen-node-defs/params.go parseCurveParams) reads a plain
// ast.BasicLit only — it has no constant-expression evaluator (that fallback exists on
// the separate ShadingParam* extractor only, a different generated file for a different
// vocabulary). math.TestVectorAngleStepIsPiOver12 (curve_params_test.go) pins this
// literal against math.Pi/12 so the two can never silently diverge.
const CurveParamVectorAngleStep = 0.2617993877991494

// vec3 aliases wire.Vec3 — the vector type and its methods (Sub/Add/Scale/
// Length/Normalize/Dot/Cross) live in nodes/wire/geometry.go now; Wiring keeps
// this name so the many existing call sites read the same.
type vec3 = wire.Vec3

// wireSegment aliases wire.WireSegment — see nodes/wire/geometry.go.
type wireSegment = wire.WireSegment
