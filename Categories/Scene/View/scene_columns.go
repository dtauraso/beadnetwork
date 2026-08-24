package View

import (
	"github.com/dtauraso/beadnetwork/Categories/Scene/Camera"
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
	ui.NodeRingPoints.Arm(sceneRoot)
	ui.BeadRingPoints.Arm(sceneRoot)
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
	if w := ui.NodeRingPoints.W(); w != nil {
		w.Begin()
		w.Surface(nodePts)
		if err := w.Flush(); err != nil {
			LogPersistErr("node_ring_point_values", "", err)
		}
	}
	if w := ui.BeadRingPoints.W(); w != nil {
		w.Begin()
		w.Surface(beadPts)
		if err := w.Flush(); err != nil {
			LogPersistErr("bead_ring_point_values", "", err)
		}
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
