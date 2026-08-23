package Startup

import (
	Speed "github.com/dtauraso/wirefold/Categories/Speed"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Scene/Scenes"

	"github.com/dtauraso/wirefold/Categories/Scene/View"
)

func InstallSpeed(ui *View.UIState, topologyPath string, speedSinks SliderPanel.Sinks) {
	speed, _ := Speed.LoadSceneSpeed(Scenes.SpeedFilePath(topologyPath))
	ui.ClockDivisor = Scenes.For(topologyPath).ClockDivisor
	ui.Speed = speed
	SliderPanel.Broadcast(speedSinks, Speed.SliderNum(speed), int64(ui.ClockDivisor))
	ui.EmitViewFrame(nil)
}
