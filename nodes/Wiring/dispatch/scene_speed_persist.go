// scene_speed_persist.go — MoveDispatch-facing side of Go's OWN playback-speed multiplier
// persister (view/speed.json): the speedPersister type, HumanEditSpeed, SliderSpeed, and
// LoadSpeed. Pure read/write helpers (WriteSceneSpeed/LoadSceneSpeed/EffectiveClockSpeed/
// BroadcastSpeed/DefaultPlaybackSpeed) live in
// nodes/Wiring/scenepersist/scene_speed_persist.go.
//
// OWNER: the view-owner goroutine (RunStdinReader, stdin_reader.go) is the SOLE caller of
// the speed Persister's Schedule() below — the clock/speed edit handler (clockAttrHandlers, in
// stdin_reader.go) is the only trigger. speed.json is scene-level and genuinely singular
// (there is only one playback speed), so it stays one file with one named owning goroutine
// (.claude/rules/persistence-ownership.md "The owner writes, and owns the path") — same
// shape as camera.json/overlays.json/sphere.json, not a per-entity split.
package dispatch

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/scene"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

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
// Nothing about the user's chosen speed changes: md.ui.speed and view/speed.json still
// hold the slider's number untouched. Only what is BROADCAST to the clocks differs, and
// only until the next start or reset.
const HumanEditSpeed = 1.0

// The speed's own file persister (view/speed.json) is one instantiation of
// scenepersist.Persister[float64] (the shared debounce-then-write actor shape, see that
// type's own doc comment), bound to WriteSceneSpeed, constructed in move_persist.go's
// EnableEditPersist and held at md.persist.speed. Armed by EnableEditPersist, then called
// exclusively by the view-owner goroutine (RunStdinReader) — see the OWNER note above. Its
// Path == "" (tests that never arm) → Schedule is a no-op.

// SliderSpeed is what the clocks run at when nothing is overriding them: the user's own
// chosen number scaled by this scene's divisor. Reading it off md.UI means the restore after
// a tilt edit can never disagree with what a live slider change would have sent.
func (md *MoveDispatch) SliderSpeed() float64 {
	return scenepersist.EffectiveClockSpeed(md.UI.Speed, md.UI.ClockDivisor)
}

// LoadSpeed reads the persisted playback-speed multiplier from speed.json (FILE DATA) into
// md.ui.speed, broadcasts it to every clock-owning goroutine's own speed channel (the SAME
// Delivery path a live slider edit uses — see clockAttrHandlers's "speed" case), and emits
// it so the buffer reflects the loaded speed from the first frame. Call after LoadTopology
// (which builds MoveDispatch and returns speedSinks) and BEFORE EnableEditPersist so this
// emit does not write the loaded/default speed back.
func (md *MoveDispatch) LoadSpeed(topologyPath string, speedSinks []chan float64, tr *T.Trace) {
	speed, _ := scenepersist.LoadSceneSpeed(scenepaths.SpeedFilePath(topologyPath))
	md.UI.ClockDivisor = scene.SceneClockDivisor(topologyPath)
	md.UI.Speed = speed
	effective := scenepersist.EffectiveClockSpeed(speed, md.UI.ClockDivisor)
	for _, ch := range speedSinks {
		clock.SendSpeedNonBlocking(ch, effective)
	}
	md.UI.EmitViewFrame(nil)
}
