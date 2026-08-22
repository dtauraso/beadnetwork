package viewstate

import (
	"github.com/dtauraso/wirefold/Categories/Scene/Camera"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/FitButton"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/NodesDropdown"
	"github.com/dtauraso/wirefold/Categories/Chrome/Tabs"
	"github.com/dtauraso/wirefold/Categories/RingPoint"
	"github.com/dtauraso/wirefold/Categories/Scene"
)

func (ui *UIState) SetSceneRoot(sceneRoot string) {
	if sceneRoot == "" {
		return
	}
	ui.sceneRoot = sceneRoot
	ui.rulesValues = PolarRulesPanel.NewValueWriter(sceneRoot)
	ui.nodesPillValues = NodesDropdown.NewValueWriter(sceneRoot)
	ui.anglePillValues = AngleDropdown.NewValueWriter(sceneRoot)
	ui.tabStripValues = Tabs.NewValueWriter(sceneRoot)
	ui.fitChipValues = FitButton.NewValueWriter(sceneRoot)
	ui.tiltPanelValues = TiltPanel.NewValueWriter(sceneRoot)
	ui.overlaysPillValues = Pills.NewValueWriter(sceneRoot)
	ui.ringPointValues = RingPoint.NewValueWriter(sceneRoot)
	ui.pointerTargetValues = Panels.NewValueWriter(sceneRoot)
	ui.sliderPanelValues = SliderPanel.NewValueWriter(sceneRoot)
	ui.ownerCountsValues = Scene.NewCountsValueWriter(sceneRoot)
}

func (ui *UIState) writeSceneColumns() {
	w := ui.ownerCountsValues
	if w == nil {
		return
	}
	if err := w.Write(ui.OwnerCounts.Nodes, ui.OwnerCounts.Edges); err != nil {
		LogPersistErr("owner_counts_values", "", err)
	}
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
		LogPersistErr("ring_point_values", "", err)
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
		LogPersistErr("pointer_target_values", "", err)
	}
}
