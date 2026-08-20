package scenepersist

import (
	"math"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/src/Scene/scene"

	"github.com/dtauraso/wirefold/src/Scene/scenepaths"
	"github.com/dtauraso/wirefold/src/Scene/viewstate"
	"github.com/dtauraso/wirefold/src/jsonpersist"
)

const DefaultPlaybackSpeed = 1.0

func SliderSpeed(ui *viewstate.UIState) float64 {
	return EffectiveClockSpeed(ui.Speed, ui.ClockDivisor)
}

func InstallSpeed(ui *viewstate.UIState, topologyPath string, speedSinks SliderPanel.Sinks) {
	speed, _ := LoadSceneSpeed(scenepaths.SpeedFilePath(topologyPath))
	ui.ClockDivisor = scene.For(topologyPath).ClockDivisor
	ui.Speed = speed
	SliderPanel.Broadcast(speedSinks, SliderNum(speed), int64(ui.ClockDivisor))
	ui.EmitViewFrame(nil)
}

func EffectiveClockSpeed(userSpeed, divisor float64) float64 {
	if divisor <= 0 {
		return userSpeed
	}
	return userSpeed / divisor
}

func WriteSceneSpeed(speedPath string, speed float64) error {
	return jsonpersist.WriteJSONAtomic(speedPath, speed)
}

func LoadSceneSpeed(speedPath string) (float64, bool) {
	var speed float64
	if !jsonpersist.ReadJSONIfExists(speedPath, &speed) {
		return DefaultPlaybackSpeed, false
	}
	return speed, true
}

func SliderNum(userSpeed float64) int64 {
	return int64(math.Round(userSpeed * SliderPanel.NumScale))
}
