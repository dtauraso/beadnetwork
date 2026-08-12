# PairNode — pacing and clock speed

[← BEHAVIOR.md](BEHAVIOR.md)

The clock a bead's round trip is measured in (`clk.Tick()`) and the clock that PACES this
node's own cycle (`clk.SleepCycle`, `nodes/clock/clock.go`) are scaled by the scene's
playback speed: `SleepCycle` waits `pulsesPerCycle()` broadcaster pulses, `1/speed`
rounded up and clamped to `[1, 64]`, so one cycle is one SCALED tick's worth of wall time
and this goroutine itself runs slower when the slider says slower. The bead marks the
round trip; the CYCLE — how often this node's loop drains its channels and can react at
all — is what the dial actually paces.

While a ▲/▼ panel click is being applied, every clock in the scene is broadcast to
`Wiring.HumanEditSpeed` (1.0, unscaled) instead of the slider's speed
(`applyUpdateTiltVector`, `nodes/Wiring/stdinreader/dispatch_apply.go`), so a click is answered on the
next real-time cycle rather than sitting unanswered for a scaled cycle (up to ~1 second at
a slow divisor). START and RESET both restore the slider's own speed
(`md.SliderSpeed()`) before doing anything else — START because running the exchange is
exactly what the slider's number is about, RESET because setting is over. The slider's own
persisted number (view/speed.json, per topology) is untouched throughout; only what is broadcast to
the clocks changes.
