package Wiring

// quant_offset_persist.go — the WRITE side of the quantized scalar triple (a,b,c) =
// (iTheta,iPhi,iR) as file data.
//
// Path construction (positionFilePath, localPolarsFilePath) lives
// in node_mover.go, not here: those are node paths, and node_mover.go is the node's
// owning file (.claude/rules/persistence-ownership.md "The owner writes, and owns the path").
//
// A node's PERSISTED position is its EXACT scene-polar (r,θ,φ) about the scene center —
// lossless, so a dragged node reloads at exactly where it was dropped. The quantized
// scalar triple (quantITheta/quantIPhi/quantIR + steps) rides along as a self-describing
// cache of the drag-time snap cells, NOT the position source. commitNodeMoveLocal calls
// nm.persistQuantOffset (below) for the dragged node's OWN nodeMover, writing
// scenePolarR/Theta/Phi + the quant cache to `<root>/nodes/<id>/position.json` —
// one-file-per-writer: this file has exactly one writer per node id (writeQuantOffset), so
// each write is a fresh whole-file marshal, no read-modify-write, no entityFileMu
// (deleted). Static node identity (id/type/r/gate) stays in meta.json, which this write
// never touches. There is no local-polars.json any more — a node has no stored record of a
// NEIGHBOUR's coordinate (MODEL.md "the polar model"); local-polars.json and its
// reader/writer were deleted with the LocalPolar type.
//
// Go owns persistence (MODEL.md): fire-and-forget, SYNCHRONOUS — persistQuantOffset writes
// immediately, inline on the caller's own goroutine (see scene_persist.go's header comment
// for why the prior debounce was removed) — logs on error, never blocks the gesture. Only
// nm.persistRoot == "" (never armed via EnableEditPersist) is a no-op. The owning goroutine
// IS the node's own nodeMover (.claude/rules/persistence-ownership.md "The owner writes, and owns the path") —
// every call site above runs on nm's own goroutine, never routed through MoveDispatch.
//
// LEGACY FALLBACK: an existing pre-split topology has these fields inline in meta.json
// instead of a separate position.json/local-polars.json — loader_tree.go's loadTree reads
// meta.json first (still required — it owns id/type/r/gate) and then overlays position.json
// / local-polars.json when present, so an old topology still loads unchanged and the next
// drag/move writes forward into the new files without ever migrating or deleting meta.json.

import (
	"fmt"
)

// persistQuantOffset writes THIS node's own exact position (scene) plus its quantized
// triple to its OWN position.json, synchronously, on THIS node's own mover goroutine
// (commitNodeMoveLocal calls it from nm's own inbox-drain goroutine — see node_move.go).
// nm.persistRoot == "" (unarmed — bare test construction, or no EnableEditPersist call)
// makes this a no-op. scene is the authoritative persisted position.
//
// UNLIKE this package's scene-level persisters (camera/overlays/sphere, each with ONE
// owning goroutine), every node has its OWN mover goroutine, so many different nodes' own
// persistQuantOffset calls can run concurrently — safe because each writes to a DIFFERENT
// file (position.json is keyed by node id, so no two calls ever race the same
// os.WriteFile/Rename) and nm.persistRoot is set once, before any mover goroutine starts,
// and never written again.
func (nm *nodeGeometry) persistQuantOffset(off quantizedOffset, scene polar) {
	if nm.persistRoot == "" {
		return
	}
	// Carries this node's CURRENT vector-angle index along unchanged — writeQuantOffset
	// is a fresh whole-file marshal (no read-modify-write), so a position-only write here
	// must still round-trip whatever topTiltVectorThetaIdx this node already holds,
	// or a later drag would silently reset a previously-set vector direction back to 0.
	if err := writeQuantOffset(nm.persistRoot, nm.id, off, scene, nm.topTiltVectorThetaIdx); err != nil {
		logPersistErr("quant_offset_persist", nm.id, err)
	}
}

// persistTiltVectorAngle writes THIS node's own vector-direction indices to its OWN
// position.json, synchronously, on THIS node's own mover goroutine (handle's
// moveMsgKindTiltVectorAngle case). Carries the node's CURRENT position/quant-offset along
// unchanged (same one-file whole-marshal shape as persistQuantOffset, reversed: this
// write is angle-driven, not position-driven). nm.persistRoot == "" is a no-op.
func (nm *nodeGeometry) persistTiltVectorAngle() {
	if nm.persistRoot == "" {
		return
	}
	if err := writeQuantOffset(nm.persistRoot, nm.id, nm.quantOffset, nm.geom.ScenePolar, nm.topTiltVectorThetaIdx); err != nil {
		logPersistErr("quant_offset_persist", nm.id, err)
	}
}

// positionFileJSON is the shape of nodes/<id>/position.json — the node's exact scene-polar
// position plus its quantized-scalar-triple cache. Mirrors the equivalent fields of
// loader_tree.go's jsonMeta (the legacy, still-read fallback shape).
type positionFileJSON struct {
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

// writeQuantOffset writes the node's EXACT scenePolarR/Theta/Phi (the authoritative,
// lossless position — see the package doc comment above) PLUS the quantized scalar triple
// (iTheta,iPhi,iR) as a self-describing cache of the drag-time snap cells, as the WHOLE
// content of <root>/nodes/<id>/position.json — the sole writer of that file, so each write
// is a fresh marshal (no read-modify-write, and no leftover `reference` field to drop: that
// was a meta.json-only artifact of the removed reference-tree model).
func writeQuantOffset(root, id string, off quantizedOffset, scene polar, topTiltVectorThetaIdx int32) error {
	if !safeTreePathComponent(id) {
		return fmt.Errorf("unsafe node id %q", id)
	}
	t, p, r := off.effectiveSteps()
	return writeJSONAtomic(positionFilePath(root, id), positionFileJSON{
		ScenePolarR: scene.R, ScenePolarTheta: scene.Theta, ScenePolarPhi: scene.Phi,
		QuantITheta: off.iTheta, QuantIPhi: off.iPhi, QuantIR: off.iR,
		StepTheta: t, StepPhi: p, StepR: r,
		TopTiltVectorThetaIdx: topTiltVectorThetaIdx,
	})
}
