package viewstate

import (
	"encoding/binary"
	"math"

	B "github.com/dtauraso/wirefold/src/Buffer"
	"github.com/dtauraso/wirefold/src/Buffer/colstream"
	"github.com/dtauraso/wirefold/src/Camera"
	"github.com/dtauraso/wirefold/src/Polar/polar"
)

func (ui *UIState) SetSingletonColumns(set *colstream.ColumnSet) {
	ui.singletonCols = set
}

func (ui *UIState) writeSceneColumns(sc polar.SceneSphere) {
	c := ui.singletonCols
	if c == nil {
		return
	}
	c.SetF32(B.ColStreamSceneCX, float32(sc.Center.X))
	c.SetF32(B.ColStreamSceneCY, float32(sc.Center.Y))
	c.SetF32(B.ColStreamSceneCZ, float32(sc.Center.Z))
	c.SetF32(B.ColStreamSceneRadius, float32(sc.Radius))
	c.SetF32(B.ColStreamSceneConstantR, float32(ui.Constants.ConstantR))
	c.SetI32(B.ColStreamSceneMaxIndexPhi, int32(ui.Constants.MaxIndexPhi))
	c.SetI32(B.ColStreamSceneMaxIndexTheta, int32(ui.Constants.MaxIndexTheta))
	c.SetI32(B.ColStreamSceneNodeCount, ui.OwnerCounts.Nodes)
	c.SetI32(B.ColStreamSceneEdgeCount, ui.OwnerCounts.Edges)
}

func writeRingSurfaceColumns(c *colstream.ColumnSet, baseX, baseY, baseZ int, pts []float32) {
	if c == nil || len(pts)%3 != 0 {
		return
	}
	n := len(pts) / 3
	xs := make([]byte, 0, n*4)
	ys := make([]byte, 0, n*4)
	zs := make([]byte, 0, n*4)
	for i := 0; i < n; i++ {
		xs = binary.LittleEndian.AppendUint32(xs, math.Float32bits(pts[i*3]))
		ys = binary.LittleEndian.AppendUint32(ys, math.Float32bits(pts[i*3+1]))
		zs = binary.LittleEndian.AppendUint32(zs, math.Float32bits(pts[i*3+2]))
	}
	c.SetBytes(baseX, xs)
	c.SetBytes(baseY, ys)
	c.SetBytes(baseZ, zs)
}

func (ui *UIState) WriteNodeRingSurfaceColumns(pts []float32) {
	writeRingSurfaceColumns(ui.singletonCols,
		B.ColStreamNodeRingPointX, B.ColStreamNodeRingPointY, B.ColStreamNodeRingPointZ, pts)
}

func (ui *UIState) WriteBeadRingSurfaceColumns(pts []float32) {
	writeRingSurfaceColumns(ui.singletonCols,
		B.ColStreamBeadRingPointX, B.ColStreamBeadRingPointY, B.ColStreamBeadRingPointZ, pts)
}

func (ui *UIState) FovDeg() float64 {
	return Camera.FovDegForHeight(ui.ViewH)
}

func (ui *UIState) writeOverlayColumns(dragNodeRow int32) {
	c := ui.singletonCols
	if c == nil {
		return
	}
	c.SetI32(B.ColStreamOverlayDragNodeRow, dragNodeRow)
	c.SetU32(B.ColStreamOverlayEditRefused, ui.EditRefused)
	c.SetU8(B.ColStreamOverlaySceneEditable, boolU8(ui.SceneEditable))
	c.SetU32(B.ColStreamOverlaySceneKinds, ui.SceneKinds)
	c.SetF32(B.ColStreamOverlaySpeed, float32(ui.Speed))
}

func (ui *UIState) writePointerTargetColumns() {
	c := ui.singletonCols
	if c == nil {
		return
	}
	t := ui.Pointer
	c.SetF32(B.ColStreamPointerTargetX, t.Rect.X)
	c.SetF32(B.ColStreamPointerTargetY, t.Rect.Y)
	c.SetF32(B.ColStreamPointerTargetW, t.Rect.W)
	c.SetF32(B.ColStreamPointerTargetH, t.Rect.H)
	c.SetU8(B.ColStreamPointerTargetKind, uint8(t.Kind))

	tx, ty, _, _ := t.TipRect(float32(ui.ViewW))
	c.SetF32(B.ColStreamPointerTargetTipX, tx)
	c.SetF32(B.ColStreamPointerTargetTipY, ty)
	c.SetBytes(B.ColStreamPointerTargetTipText, []byte(t.Tip))
}

func (ui *UIState) writePanelColumns() {
	c := ui.singletonCols
	if c == nil {
		return
	}
	c.SetU8(B.ColStreamPanelOverlays, boolU8(ui.PN.OverlaysOpen))
	c.SetU8(B.ColStreamPanelNode, boolU8(ui.PN.NodeOpen))
	c.SetU8(B.ColStreamPanelNodeShape, boolU8(ui.PN.NodeShapeOpen))
	c.SetU8(B.ColStreamPanelNodeState, boolU8(ui.PN.NodeStateOpen))
	c.SetU8(B.ColStreamPanelNodePoles, boolU8(ui.PN.NodePolesOpen))
	c.SetU8(B.ColStreamPanelNodeRules, boolU8(ui.PN.NodeRulesOpen))
	c.SetU8(B.ColStreamPanelScene, boolU8(ui.PN.SceneOpen))
	c.SetU8(B.ColStreamPanelSceneGuides, boolU8(ui.PN.SceneGuidesOpen))
	c.SetU8(B.ColStreamPanelScenePoles, boolU8(ui.PN.ScenePolesOpen))
	c.SetU8(B.ColStreamPanelSceneVectors, boolU8(ui.PN.SceneVectorsOpen))
	c.SetU8(B.ColStreamPanelSceneLabels, boolU8(ui.PN.SceneLabelsOpen))
}
