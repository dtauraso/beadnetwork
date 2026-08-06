// scene_speed_persist.go — persist + load Go's OWN playback-speed multiplier to
// view/speed.json (writer + loader + seed, mirroring scene_overlays_persist.go).
//
// OWNER: the view-owner goroutine (RunStdinReader, stdin_reader.go) is the SOLE caller of
// speedPersister.schedule() below — the clock/speed edit handler (clockAttrHandlers, in
// stdin_reader.go) is the only trigger. speed.json is scene-level and genuinely singular
// (there is only one playback speed), so it stays one file with one named owning goroutine
// (.claude/rules/persistence-ownership.md "The owner writes, and owns the path") — same
// shape as camera.json/overlays.json/sphere.json, not a per-entity split.
//
// Persistence has one trigger: an ON-CHANGE synchronous write scheduled whenever a
// clock/speed edit lands (mirrors camera/overlays' on-change write — see
// scene_persist.go's header comment for why the old debounce was removed).
//
// LOAD side: loadSceneSpeed reads the value back; MoveDispatch.LoadSpeed installs it into
// md.ui.speed on startup, broadcasts it to every clock-owning goroutine's own speed channel
// the SAME way a live slider edit does (wire.SendSpeedNonBlocking over speedSinks), and
// emits it so the buffer reflects the loaded speed from the first frame — closing the
// slider→reload→still-at-that-speed round trip.
//
// UNLIKE counts.json, a missing or malformed speed.json is a PREFERENCE, not a structural
// invariant: it falls back to the default speed (1) quietly rather than failing loudly.

package Wiring

import (
	"encoding/json"

	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

// defaultPlaybackSpeed is the speed a fresh topology (or a missing/malformed speed.json)
// falls back to.
const defaultPlaybackSpeed = 1.0

// EffectiveClockSpeed is the ONE place userSpeed (the slider's number, unscaled, the same
// value persisted to view/speed.json) is turned into the rate actually broadcast to the
// clocks: userSpeed / scene's ClockDivisor (SceneTab.ClockDivisor / SceneClockDivisor,
// scene_tabs.go). Both broadcast sites — the live slider edit (clockAttrHandlers's "speed"
// case) and the load-time seed (LoadSpeed) — call this so they can never disagree.
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

// writeSceneSpeed writes the current playback-speed multiplier as the WHOLE content of
// speedPath (speed.json) — the sole writer of that file.
func writeSceneSpeed(speedPath string, speed float64) error {
	obj := map[string]json.RawMessage{
		"speed": json.RawMessage(formatSpeedJSON(speed)),
	}
	return writeJSONAtomic(speedPath, obj)
}

// formatSpeedJSON renders speed as a plain JSON number (no trailing zeros games needed —
// encoding/json's own float formatting is used via a round-trip through json.Marshal so
// 0.25/0.75 etc. survive exactly).
func formatSpeedJSON(speed float64) []byte {
	b, err := json.Marshal(speed)
	if err != nil {
		b = []byte("1")
	}
	return b
}

// speedPersister writes the playback speed to speed.json as it changes. Armed by
// EnableEditPersist, then called exclusively by the view-owner goroutine (RunStdinReader)
// — see the OWNER note above. path == "" (tests that never arm) → no-op.
type speedPersister struct {
	path string // speed.json path (speedFilePath(topologyPath))
}

// schedule writes the given speed to speed.json synchronously.
func (p *speedPersister) schedule(speed float64) {
	if p == nil || p.path == "" {
		return
	}
	if err := writeSceneSpeed(p.path, speed); err != nil {
		logPersistErr("scene_speed_persist", p.path, err)
		return
	}
}

// sceneSpeedFile is the on-disk shape of speed.json.
type sceneSpeedFile struct {
	Speed *float64 `json:"speed"`
}

// loadSceneSpeed reads the persisted playback speed from speedPath (speed.json). The bool
// return is false when the file yields no speed key (fresh topology, or a missing/malformed
// file) — the caller then keeps defaultPlaybackSpeed. This is a PREFERENCE, not a structural
// invariant (unlike counts.json), so a missing/malformed file falls back quietly rather than
// failing loudly — see readJSONBestEffort.
func loadSceneSpeed(speedPath string) (float64, bool) {
	var sf sceneSpeedFile
	readJSONBestEffort(speedPath, &sf)
	if sf.Speed == nil {
		return defaultPlaybackSpeed, false
	}
	return *sf.Speed, true
}

// LoadSpeed reads the persisted playback-speed multiplier from speed.json (FILE DATA) into
// md.ui.speed, broadcasts it to every clock-owning goroutine's own speed channel (the SAME
// Delivery path a live slider edit uses — see clockAttrHandlers's "speed" case), and emits
// it so the buffer reflects the loaded speed from the first frame. Call after LoadTopology
// (which builds MoveDispatch and returns speedSinks) and BEFORE EnableEditPersist so this
// emit does not write the loaded/default speed back.
func (md *MoveDispatch) LoadSpeed(topologyPath string, speedSinks []chan float64, tr *T.Trace) {
	speed, _ := loadSceneSpeed(speedFilePath(topologyPath))
	md.ui.clockDivisor = SceneClockDivisor(topologyPath)
	md.ui.speed = speed
	effective := EffectiveClockSpeed(speed, md.ui.clockDivisor)
	for _, ch := range speedSinks {
		wire.SendSpeedNonBlocking(ch, effective)
	}
	md.emitViewFrame(nil)
}
