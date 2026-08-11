// scene_speed_persist.go — pure read/write helpers + the divisor arithmetic for Go's OWN
// playback-speed multiplier in view/speed.json, plus HumanEditSpeed, SliderSpeed, and
// InstallSpeed (the view-owner-goroutine-facing entry point that seeds ui.Speed from disk,
// broadcasts it, and emits it) — these were the former dispatch.MoveDispatch-facing methods,
// moved here since they read/wrote nothing but *viewstate.UIState
// (docs/planning/movedispatch-decomposition.md §34). The speed Persister instance itself
// still lives in nodes/Wiring/viewpersist (Persisters.speed, armed by EnableEditPersist).
//
// UNLIKE counts.json, a missing or malformed speed.json is a PREFERENCE, not a structural
// invariant: it falls back to DefaultPlaybackSpeed quietly rather than failing loudly.
package scenepersist

import (
	"encoding/json"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

// DefaultPlaybackSpeed is the speed a fresh topology (or a missing/malformed speed.json)
// falls back to.
const DefaultPlaybackSpeed = 1.0

// HumanEditSpeed is the speed every clock runs at WHILE A TILT IS BEING SET from the angle
// panel — unscaled, the same rate the ring runs at.
//
// Setting an angle is an interaction, not a simulation. Since SleepCycle became
// speed-scaled (nodes/wire/clock.go), a scene's own divisor stretches every paced loop:
// at the pair's divisor one cycle is about a second, and a node drains its TiltEditIn in
// that loop — so a ▲/▼ click could sit unanswered for a second and the panel felt dead.
// The slider's number is about how fast the exchange should RUN, and it should not be
// what decides how fast a click is noticed.
//
// The boundaries are the user's own actions rather than a timer or an idle guess: an angle
// click switches to this speed, and START or RESET puts the slider's speed back
// (applyUpdateTiltVector). Those are exactly the edges of "setting the tilt" — you stop
// setting when you start the exchange, or when you abandon it.
//
// Nothing about the user's chosen speed changes: ui.Speed and view/speed.json still
// hold the slider's number untouched. Only what is BROADCAST to the clocks differs, and
// only until the next start or reset.
const HumanEditSpeed = 1.0

// SliderSpeed is what the clocks run at when nothing is overriding them: the user's own
// chosen number scaled by this scene's divisor. Reading it off ui means the restore after
// a tilt edit can never disagree with what a live slider change would have sent.
func SliderSpeed(ui *viewstate.UIState) float64 {
	return EffectiveClockSpeed(ui.Speed, ui.ClockDivisor)
}

// InstallSpeed reads the persisted playback-speed multiplier from speed.json (FILE DATA) into
// ui.Speed, broadcasts it to every clock-owning goroutine's own speed channel (the SAME
// Delivery path a live slider edit uses — see clockAttrHandlers's "speed" case), and emits
// it so the buffer reflects the loaded speed from the first frame. Call after LoadTopology
// (which builds MoveDispatch and returns speedSinks) and BEFORE EnableEditPersist so this
// emit does not write the loaded/default speed back.
func InstallSpeed(ui *viewstate.UIState, topologyPath string, speedSinks []chan float64, tr *T.Trace) {
	speed, _ := LoadSceneSpeed(scenepaths.SpeedFilePath(topologyPath))
	ui.ClockDivisor = scene.SceneClockDivisor(topologyPath)
	ui.Speed = speed
	effective := EffectiveClockSpeed(speed, ui.ClockDivisor)
	BroadcastSpeed(speedSinks, effective)
	ui.EmitViewFrame(nil)
}

// EffectiveClockSpeed is the ONE place userSpeed (the slider's number, unscaled, the same
// value persisted to view/speed.json) is turned into the rate actually broadcast to the
// clocks: userSpeed / scene's ClockDivisor (SceneTab.ClockDivisor / SceneClockDivisor,
// scene/scene_tabs.go). Both broadcast sites — the live slider edit (clockAttrHandlers's "speed"
// case) and the load-time seed (MoveDispatch.LoadSpeed) — call this so they can never disagree.
//
// divisor is guarded against 0/negative (SceneClockDivisor should never return one, but a
// division must never be reachable even if that guarantee is ever violated) — such a value
// is treated as "no scaling" (divisor 1) rather than dividing.
func EffectiveClockSpeed(userSpeed, divisor float64) float64 {
	if divisor <= 0 {
		return userSpeed
	}
	return userSpeed / divisor
}

// WriteSceneSpeed writes the current playback-speed multiplier as the WHOLE content of
// speedPath (speed.json) — the sole writer of that file.
func WriteSceneSpeed(speedPath string, speed float64) error {
	obj := map[string]json.RawMessage{
		"speed": json.RawMessage(FormatSpeedJSON(speed)),
	}
	return jsonpersist.WriteJSONAtomic(speedPath, obj)
}

// FormatSpeedJSON renders speed as a plain JSON number (no trailing zeros games needed —
// encoding/json's own float formatting is used via a round-trip through json.Marshal so
// 0.25/0.75 etc. survive exactly).
func FormatSpeedJSON(speed float64) []byte {
	b, err := json.Marshal(speed)
	if err != nil {
		b = []byte("1")
	}
	return b
}

// sceneSpeedFile is the on-disk shape of speed.json.
type sceneSpeedFile struct {
	Speed *float64 `json:"speed"`
}

// LoadSceneSpeed reads the persisted playback speed from speedPath (speed.json). The bool
// return is false when the file yields no speed key (fresh topology, or a missing/malformed
// file) — the caller then keeps DefaultPlaybackSpeed. This is a PREFERENCE, not a structural
// invariant (unlike counts.json), so a missing/malformed file falls back quietly rather than
// failing loudly — see jsonpersist.ReadJSONBestEffort.
func LoadSceneSpeed(speedPath string) (float64, bool) {
	var sf sceneSpeedFile
	jsonpersist.ReadJSONBestEffort(speedPath, &sf)
	if sf.Speed == nil {
		return DefaultPlaybackSpeed, false
	}
	return *sf.Speed, true
}

// BroadcastSpeed sends one effective speed to every clock-owning goroutine's own channel —
// the SAME Delivery path MoveDispatch.LoadSpeed and a live slider edit use, factored out so
// the tilt-edit speed override (HumanEditSpeed, Wiring package) cannot drift from it.
func BroadcastSpeed(speedSinks []chan float64, effective float64) {
	for _, ch := range speedSinks {
		clock.SendSpeedNonBlocking(ch, effective)
	}
}
