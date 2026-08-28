package View

import (
	"fmt"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/AngleDropdown"
)

func (ui *UIState) SetViewport(w, h float64) {
	changed := w != ui.ViewW || h != ui.ViewH
	ui.ViewW = w
	ui.ViewH = h
	if !changed {
		return
	}
	ui.EmitBreadcrumb(RowEvent{
		Label: BreadcrumbViewport, NodeRow: -1, PortRow: -1, TargetRow: -1,
		TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Value: int32(ui.FovDeg()),
		Text:  fmt.Sprintf("%.4fx%.4f", w, h),
	})
}

func (ui *UIState) SetLatticePoints(points int32, persist func(int32), broadcast func(int32)) {
	if points < AngleDropdown.LatticePointsMin || points > AngleDropdown.LatticePointsMax || points%4 != 0 {
		return
	}
	ui.LatticePoints = points
	persist(points)
	broadcast(points)
}
