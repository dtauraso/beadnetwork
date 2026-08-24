package Startup

import (
	Speed "github.com/dtauraso/beadnetwork/Categories/Speed"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"

	"github.com/dtauraso/beadnetwork/Categories/Scene/View"
)

func InstallSpeed(ui *View.UIState, topologyPath string, speedSinks SliderPanel.Sinks) {
	speed, _ := Speed.LoadSceneSpeed(Scenes.SpeedFilePath(topologyPath))
	ui.ClockDivisor = Scenes.For(topologyPath).ClockDivisor
	ui.Speed = speed
	SliderPanel.Broadcast(speedSinks, Speed.SliderNum(speed), int64(ui.ClockDivisor))
	ui.EmitViewFrame(nil)
}
