package viewstate

import (
	B "github.com/dtauraso/wirefold/src/Buffer"
	"github.com/dtauraso/wirefold/src/Buffer/colstream"
	"github.com/dtauraso/wirefold/src/Camera"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/NodesDropdown"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/FitButton"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/src/RingPoint"
	"github.com/dtauraso/wirefold/src/Chrome/Panels"
	"github.com/dtauraso/wirefold/src/valuefile"
	"github.com/dtauraso/wirefold/src/Chrome/Pills"
	"github.com/dtauraso/wirefold/src/Chrome/Tabs"
)

func (ui *UIState) SetSingletonColumns(set *colstream.ColumnSet) {
	ui.singletonCols = set
}

func (ui *UIState) SetSceneRoot(sceneRoot string) {
	if sceneRoot == "" {
		return
	}
	ui.rulesValues = PolarRulesPanel.NewValueWriter(sceneRoot)
	ui.nodesPillValues = NodesDropdown.NewValueWriter(sceneRoot)
	ui.anglePillValues = AngleDropdown.NewValueWriter(sceneRoot)
	ui.tabStripValues = Tabs.NewValueWriter(sceneRoot)
	ui.fitChipValues = FitButton.NewValueWriter(sceneRoot)
	ui.tiltPanelValues = TiltPanel.NewValueWriter(sceneRoot)
	ui.overlaysPillValues = Pills.NewValueWriter(sceneRoot)
	ui.ringPointValues = RingPoint.NewValueWriter(sceneRoot)
	ui.pointerTargetValues = Panels.NewValueWriter(sceneRoot)
}

func (ui *UIState) writeSceneColumns() {
	c := ui.singletonCols
	if c == nil {
		return
	}
	c.SetI32(B.ColStreamSceneNodeCount, ui.OwnerCounts.Nodes)
	c.SetI32(B.ColStreamSceneEdgeCount, ui.OwnerCounts.Edges)
}

func (ui *UIState) WriteRingSurfaces(nodePts, beadPts []float32) {
	w := ui.ringPointValues
	if w == nil {
		return
	}
	w.Begin()
	w.Surface("nodeX", "nodeY", "nodeZ", nodePts)
	w.Surface("beadX", "beadY", "beadZ", beadPts)
	if err := w.Flush(); err != nil {
		valuefile.LogPersistErr("ring_point_values", "", err)
	}
}

func (ui *UIState) FovDeg() float64 {
	return Camera.FovDegForHeight(ui.ViewH)
}

func (ui *UIState) writePointerTargetColumns() {
	w := ui.pointerTargetValues
	if w == nil {
		return
	}
	t := ui.Pointer
	tx, ty, _, _ := t.TipRect(float32(ui.ViewW))
	if err := w.Write(t.Rect.X, t.Rect.Y, t.Rect.W, t.Rect.H, uint8(t.Kind), tx, ty, t.Tip); err != nil {
		valuefile.LogPersistErr("pointer_target_values", "", err)
	}
}

