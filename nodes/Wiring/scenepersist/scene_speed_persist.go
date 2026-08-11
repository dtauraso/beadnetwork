// scene_speed_persist.go — pure read/write helpers + the divisor arithmetic for Go's OWN
// playback-speed multiplier in view/speed.json, lifted out of
// nodes/Wiring/scene_speed_persist.go (which keeps the MoveDispatch-facing methods,
// HumanEditSpeed, and the speedPersister type — see that file's own header).
//
// UNLIKE counts.json, a missing or malformed speed.json is a PREFERENCE, not a structural
// invariant: it falls back to DefaultPlaybackSpeed quietly rather than failing loudly.
package scenepersist

import (
	"encoding/json"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

// DefaultPlaybackSpeed is the speed a fresh topology (or a missing/malformed speed.json)
// falls back to.
const DefaultPlaybackSpeed = 1.0

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
