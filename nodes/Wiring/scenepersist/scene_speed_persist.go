package scenepersist

import (
	"encoding/json"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	"github.com/dtauraso/wirefold/nodes/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

const DefaultPlaybackSpeed = 1.0

const HumanEditSpeed = 1.0

func SliderSpeed(ui *viewstate.UIState) float64 {
	return EffectiveClockSpeed(ui.Speed, ui.ClockDivisor)
}

func InstallSpeed(ui *viewstate.UIState, topologyPath string, speedSinks []chan float64, tr *T.Trace) {
	speed, _ := LoadSceneSpeed(scenepaths.SpeedFilePath(topologyPath))
	ui.ClockDivisor = scene.SceneClockDivisor(topologyPath)
	ui.Speed = speed
	effective := EffectiveClockSpeed(speed, ui.ClockDivisor)
	BroadcastSpeed(speedSinks, effective)
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

func BroadcastSpeed(speedSinks []chan float64, effective float64) {
	for _, ch := range speedSinks {
		clock.SendSpeedNonBlocking(ch, effective)
	}
}
