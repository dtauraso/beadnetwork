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

// vec3 aliases wire.Vec3 — the vector type and its methods (Sub/Add/Scale/
// Length/Normalize/Dot/Cross) live in nodes/wire/geometry.go now; Wiring keeps
// this name so the many existing call sites read the same.
type vec3 = wire.Vec3

// wireSegment aliases wire.WireSegment — see nodes/wire/geometry.go.
type wireSegment = wire.WireSegment
