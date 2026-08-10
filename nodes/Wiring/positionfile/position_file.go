// Package positionfile is the schema and path for nodes/<id>/position.json — a node's
// exact scene-polar position plus its quantized-scalar-triple cache
// (nodes/Wiring/quant_offset_persist.go's doc comment has the full model: scenePolarR/
// Theta/Phi is the authoritative, lossless position; the quant fields are a
// self-describing cache of the drag-time snap cells, not the position source).
//
// It exists as its own package, separate from nodes/Wiring, so both the writer
// (nodes/Wiring's own node_mover.go/quant_offset_persist.go, the node's own mover — see
// .claude/rules/persistence-ownership.md "The owner writes, and owns the path") and the
// reader (nodes/Wiring/loadspec's loader_tree.go) can import one shared schema+path
// definition instead of loadspec depending on nodes/Wiring for it — that dependency is
// exactly the cycle nodes/Wiring/loadspec's lift needed to avoid.
package positionfile

import "path/filepath"

// FilePath is <root>/nodes/<id>/position.json.
func FilePath(root, id string) string {
	return filepath.Join(root, "nodes", id, "position.json")
}

// JSON is the shape of nodes/<id>/position.json — the node's exact scene-polar position
// plus its quantized-scalar-triple cache. Mirrors the equivalent fields of
// loadspec's loader_tree.go's jsonMeta (the legacy, still-read fallback shape).
type JSON struct {
	ScenePolarR     float64 `json:"scenePolarR"`
	ScenePolarTheta float64 `json:"scenePolarTheta"`
	ScenePolarPhi   float64 `json:"scenePolarPhi"`
	QuantITheta     int     `json:"quantITheta"`
	QuantIPhi       int     `json:"quantIPhi"`
	QuantIR         int     `json:"quantIR"`
	StepTheta       float64 `json:"stepTheta"`
	StepPhi         float64 `json:"stepPhi"`
	StepR           float64 `json:"stepR"`
	// TopTiltVectorThetaIdx is this node's own vector-direction index
	// (node_mover.go's topTiltVectorThetaIdx — an integer count of
	// TiltVectorAngleStep, never a stored float). Omitted (0 = world +y) for a topology
	// saved before this field existed — matches the pre-existing hardcoded +y default.
	// There is no φ counterpart any more (task/drop-tilt-vector-phi); a file saved by an
	// older build that still carries "topTiltVectorPhiIdx" simply has that key ignored on
	// load — encoding/json drops unrecognized fields by default, so an old file is not a
	// load error.
	TopTiltVectorThetaIdx int32 `json:"topTiltVectorThetaIdx,omitempty"`
}
