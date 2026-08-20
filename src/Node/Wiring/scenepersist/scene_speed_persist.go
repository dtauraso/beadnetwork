package scenepersist

import (
	"github.com/dtauraso/wirefold/src/Node/Wiring/scene"
	"github.com/dtauraso/wirefold/src/Chrome/SliderPanel"
	"math"

	"encoding/json"

	"github.com/dtauraso/wirefold/src/Node/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/src/Node/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/src/Node/Wiring/viewstate"

	T "github.com/dtauraso/wirefold/src/Trace"
)

const DefaultPlaybackSpeed = 1.0

func SliderSpeed(ui *viewstate.UIState) float64 {
	return EffectiveClockSpeed(ui.Speed, ui.ClockDivisor)
}

func InstallSpeed(ui *viewstate.UIState, topologyPath string, speedSinks SliderPanel.Sinks, tr *T.Trace) {
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
	obj := map[string]json.RawMessage{
		"speed": json.RawMessage(FormatSpeedJSON(speed)),
	}
	return jsonpersist.WriteJSONAtomic(speedPath, obj)
}

func FormatSpeedJSON(speed float64) []byte {
	b, err := json.Marshal(speed)
	if err != nil {
		b = []byte("1")
	}
	return b
}

type sceneSpeedFile struct {
	Speed *float64 `json:"speed"`
}

func LoadSceneSpeed(speedPath string) (float64, bool) {
	var sf sceneSpeedFile
	jsonpersist.ReadJSONBestEffort(speedPath, &sf)
	if sf.Speed == nil {
		return DefaultPlaybackSpeed, false
	}
	return *sf.Speed, true
}

func SliderNum(userSpeed float64) int64 {
	return int64(math.Round(userSpeed * SliderPanel.NumScale))
}
