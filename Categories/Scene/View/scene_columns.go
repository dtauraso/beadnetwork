package View

import (
	"github.com/dtauraso/wirefold/Categories/Scene/Camera"
)

func (ui *UIState) SetSceneRoot(sceneRoot string) {
	if sceneRoot == "" {
		return
	}
	ui.sceneRoot = sceneRoot
	ui.Rules.Arm(sceneRoot)
	ui.Nodes.Arm(sceneRoot)
	ui.Angle.Arm(sceneRoot)
	ui.TabStrip.Arm(sceneRoot)
	ui.Fit.Arm(sceneRoot)
	ui.Tilt.Arm(sceneRoot)
	ui.OverlaysPill.Arm(sceneRoot)
	ui.RingPoints.Arm(sceneRoot)
	ui.PointerBlk.Arm(sceneRoot)
	ui.Slider.Arm(sceneRoot)
	ui.Counts.Arm(sceneRoot)
}

func (ui *UIState) writeSceneColumns() {
	if err := ui.Counts.Write(ui.OwnerCounts.Nodes, ui.OwnerCounts.Edges); err != nil {
		LogPersistErr("owner_counts_values", "", err)
	}
}

func (ui *UIState) WriteRingSurfaces(nodePts, beadPts []float32) {
	w := ui.RingPoints.W()
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
	w := ui.PointerBlk.W()
	if w == nil {
		return
	}
	t := ui.Pointer
	tx, ty, _, _ := t.TipRect(float32(ui.ViewW))
	if err := w.Write(t.Rect.X, t.Rect.Y, t.Rect.W, t.Rect.H, uint8(t.Kind), tx, ty, t.Tip); err != nil {
		LogPersistErr("pointer_target_values", "", err)
	}
}
