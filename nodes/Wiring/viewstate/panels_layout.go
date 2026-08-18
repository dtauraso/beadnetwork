package viewstate

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/panelstack"
	"github.com/dtauraso/wirefold/nodes/Wiring/speedpanel"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltpanel"
)

type PanelLayout struct {
	Speed speedpanel.Layout
	Tilt  tiltpanel.Layout
}

func (ui *UIState) PanelLayout() PanelLayout {
	st := panelstack.New()
	return PanelLayout{
		Speed: speedpanel.Build(st),
		Tilt:  tiltpanel.Build(st, ui.TiltRows, ui.TiltLabels),
	}
}
