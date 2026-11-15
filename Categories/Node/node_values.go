package Node

import (
	"fmt"
	"path/filepath"
)

const ValueRelTemplate = "view/nodes/{row}/node.bin"

var NodeValueNames = buildNodeValueNames()

func buildNodeValueNames() []string {
	names := []string{
		"radius", "navTubeR",
		"poleAnchorX", "poleAnchorY", "poleAnchorZ",
		"labelAnchorX", "labelAnchorY", "labelAnchorZ",
		"polePhi", "poleTheta", "poleRingR",
	}
	for m := range 16 {
		names = append(names, fmt.Sprintf("ringM%d", m))
	}
	for m := range 16 {
		names = append(names, fmt.Sprintf("bodyM%d", m))
	}
	return append(names,
		"selected", "kindId", "label", "latchedSel",
	)
}

func RingName(m int) string { return fmt.Sprintf("ringM%d", m) }

func BodyName(m int) string { return fmt.Sprintf("bodyM%d", m) }

func ValueRelPath(row int) string {
	return fmt.Sprintf("view/nodes/%d/node.bin", row)
}

type ValueWriter struct {
	*BlobWriter
}

func NewValueWriter(sceneRoot string, row int) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelPath(row)))
	return &ValueWriter{BlobWriter: NewBlobWriter(path, NodeBlockValueNames)}
}

func WriteNodeValues(w *ValueWriter, f NodeState) {
	if w == nil {
		return
	}
	w.F32("radius", f.Radius)

	w.F32("navTubeR", f.NavTubeR)
	w.F32("poleAnchorX", f.PoleAnchorX)
	w.F32("poleAnchorY", f.PoleAnchorY)
	w.F32("poleAnchorZ", f.PoleAnchorZ)

	w.F32("labelAnchorX", f.LabelAnchorX)
	w.F32("labelAnchorY", f.LabelAnchorY)
	w.F32("labelAnchorZ", f.LabelAnchorZ)

	w.F32("polePhi", f.PolePhi)
	w.F32("poleTheta", f.PoleTheta)
	w.F32("poleRingR", f.PoleRingR)

	for m := range 16 {
		w.F32(RingName(m), f.RingMatrix[m])
	}
	for m := range 16 {
		w.F32(BodyName(m), f.BodyMatrix[m])
	}

	w.U8("selected", f.Selected)
	w.U8("kindId", f.KindID)
	w.Text("label", f.Label)
	w.U8("latchedSel", f.LatchedSel)
}
