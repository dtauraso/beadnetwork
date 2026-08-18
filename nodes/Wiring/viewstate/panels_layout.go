package viewstate

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/angledropdown"
	"github.com/dtauraso/wirefold/nodes/Wiring/panelstack"
	"github.com/dtauraso/wirefold/nodes/Wiring/speedpanel"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltpanel"
)

type PanelLayout struct {
	Speed speedpanel.Layout
	Tilt  tiltpanel.Layout
	Angle angledropdown.Layout
}

var PillLabels = []string{angledropdown.Label, "Nodes", "Overlays"}

func (ui *UIState) PanelLayout() PanelLayout {
	st := panelstack.New()
	pills := panelstack.NewPillStack(float32(ui.ViewW), PillLabels)

	nodes := make([]angledropdown.Node, len(ui.TiltRows))
	for i, row := range ui.TiltRows {
		nodes[i] = angledropdown.Node{
			Row:   row,
			Label: ui.TiltLabels[i],
			Open:  ui.AngleGroupOpen[row],
		}
	}

	return PanelLayout{
		Speed: speedpanel.Build(st),
		Tilt:  tiltpanel.Build(st, ui.TiltRows, ui.TiltLabels),
		Angle: angledropdown.Build(pills, ui.AngleOpen, ui.LatticePoints, nodes),
	}
}
