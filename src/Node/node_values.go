package Node

import (
	"fmt"
	"path/filepath"

	"github.com/dtauraso/wirefold/src/Polar/polarindex"
	"github.com/dtauraso/wirefold/src/valuefile"
)

const ValueRelTemplate = "view/nodes/{row}/node.bin"

var NodeValueNames = buildNodeValueNames()

func buildNodeValueNames() []string {
	names := []string{
		"indexR", "indexPhi", "indexTheta",
		"hasPos", "radius", "navTubeR",
		"poleAnchorX", "poleAnchorY", "poleAnchorZ",
		"labelAnchorX", "labelAnchorY", "labelAnchorZ",
		"polePhi", "poleTheta", "poleRingR",
	}
	for m := range 16 {
		names = append(names, fmt.Sprintf("ringM%d", m))
	}
	return append(names,
		"topTiltVectorText",
		"selected", "kindId", "label", "hovered", "latchedSel",
		"roundsToParallel", "msgsToParallel",
		"dragRLocked", "dragPhiLocked", "dragThetaMax", "dragActive", "kindRuleActive",
		"selfRLocked", "selfPhiLocked", "selfThetaMax", "selfActive",
		"ruleGroupId", "ruleGroupSize",
	)
}

func RingName(m int) string { return fmt.Sprintf("ringM%d", m) }

func ValueRelPath(row int) string {
	return fmt.Sprintf("view/nodes/%d/node.bin", row)
}

type ValueWriter struct {
	*valuefile.BlobWriter
}

func NewValueWriter(sceneRoot string, row int) *ValueWriter {
	path := filepath.Join(sceneRoot, filepath.FromSlash(ValueRelPath(row)))
	return &ValueWriter{BlobWriter: valuefile.NewBlobWriter(path, NodeValueNames)}
}

func WriteNodeValues(w *ValueWriter, f NodeState) error {
	if w == nil {
		return nil
	}
	w.Begin()
	w.I32("indexR", f.IndexR)
	w.I32("indexPhi", f.IndexPhi)
	w.I32("indexTheta", f.IndexTheta)
	w.U8("hasPos", f.HasPos)
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

	w.Text("topTiltVectorText", polarindex.AngleText(f.TopTiltVectorIdx, int32(f.LatticePoints)))

	w.U8("selected", f.Selected)
	w.U8("kindId", f.KindID)
	w.Text("label", f.Label)
	w.U8("hovered", f.Hovered)
	w.U8("latchedSel", f.LatchedSel)

	w.I32("roundsToParallel", f.RoundsToParallel)
	w.I32("msgsToParallel", f.MsgsToParallel)

	w.U8("dragRLocked", f.DragRLocked)
	w.U8("dragPhiLocked", f.DragPhiLocked)
	w.F32("dragThetaMax", f.DragThetaMax)
	w.U8("dragActive", f.DragActive)
	w.U8("kindRuleActive", f.KindRuleActive)

	w.U8("selfRLocked", f.SelfRLocked)
	w.U8("selfPhiLocked", f.SelfPhiLocked)
	w.F32("selfThetaMax", f.SelfThetaMax)
	w.U8("selfActive", f.SelfActive)

	w.I32("ruleGroupId", f.RuleGroupID)
	w.I32("ruleGroupSize", f.RuleGroupSize)

	return w.Flush()
}
