package viewstate

import (
	"encoding/binary"
	"math"

	"github.com/dtauraso/wirefold/src/Node/Wiring/geom/camera"
	"github.com/dtauraso/wirefold/src/Node/Wiring/geom/polar"
	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
	"github.com/dtauraso/wirefold/src/schema/buffer-layout/colstream"
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

func (ui *UIState) writeCameraColumns() {
	c := ui.singletonCols
	if c == nil {
		return
	}
	v := ui.VP.Viewpoint
	c.SetF32(B.ColStreamCameraPX, float32(v.Pivot.X))
	c.SetF32(B.ColStreamCameraPY, float32(v.Pivot.Y))
	c.SetF32(B.ColStreamCameraPZ, float32(v.Pivot.Z))
	c.SetF32(B.ColStreamCameraR, float32(v.R))
	c.SetF32(B.ColStreamCameraPosPhi, float32(v.Pos.Phi))
	c.SetF32(B.ColStreamCameraPosTheta, float32(v.Pos.Theta))
	c.SetF32(B.ColStreamCameraUpPhi, float32(v.Up.Phi))
	c.SetF32(B.ColStreamCameraUpTheta, float32(v.Up.Theta))
	c.SetF32(B.ColStreamCameraFocalPx, float32(camera.FocalPixels))
}

func (ui *UIState) FovDeg() float64 {
	return camera.FovDegForHeight(ui.ViewH)
}

func (ui *UIState) writeOverlayColumns(dragNodeRow int32) {
	c := ui.singletonCols
	if c == nil {
		return
	}
	c.SetU8(B.ColStreamOverlaySceneTori, boolU8(ui.OV.SceneToriVisible))
	c.SetU8(B.ColStreamOverlayScenePoles, boolU8(ui.OV.ScenePolesVisible))
	c.SetU8(B.ColStreamOverlayNodePoles, boolU8(ui.OV.NodePolesVisible))
	c.SetU8(B.ColStreamOverlayHandholds, boolU8(ui.OV.HandholdsVisible))
	c.SetU8(B.ColStreamOverlayLabelsGlobal, boolU8(ui.OV.LabelsGlobalVisible))
	c.SetU8(B.ColStreamOverlayOverlaysVis, boolU8(ui.OV.OverlaysVisible))
	c.SetU8(B.ColStreamOverlayNodeBody, boolU8(ui.OV.NodeBodyVisible))
	c.SetU8(B.ColStreamOverlayNodeRing, boolU8(ui.OV.NodeRingVisible))
	c.SetU8(B.ColStreamOverlayRingPick, boolU8(ui.OV.RingPickVisible))
	c.SetU8(B.ColStreamOverlaySelectionRing, boolU8(ui.OV.SelectionRingVisible))
	c.SetU8(B.ColStreamOverlayHoverRing, boolU8(ui.OV.HoverRingVisible))
	c.SetU8(B.ColStreamOverlaySceneVectors, boolU8(ui.OV.SceneVectorsVisible))
	c.SetU8(B.ColStreamOverlayRuleChannels, boolU8(ui.OV.RuleChannelsVisible))
	c.SetU8(B.ColStreamOverlayNodePoleSphere, boolU8(ui.OV.NodePoleSphereVisible))
	c.SetU8(B.ColStreamOverlayAllPoleSpheres, boolU8(ui.OV.AllPoleSpheresVisible))

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
