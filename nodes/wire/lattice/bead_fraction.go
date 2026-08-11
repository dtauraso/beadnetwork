// bead_fraction.go — the pure per-bead progress math shared by everywhere a wire
// reports "how far across has this bead got" (nodes/wire's live_beads.go,
// bead_advance.go's advanceBead, and paced_wire_drive.go's
// ReviseInFlightGeometry): a bead's fractional progress t
// (0..1) is always the same clamp-and-divide of (nowTick, placementTick,
// crossTicks), independent of any wire's own in-flight state — lifted here so the
// three call sites share one definition instead of three copies of the same
// clamp. Takes and returns only float64; it never reads or holds a bead or wire.
package lattice

import "github.com/dtauraso/wirefold/nodes/wire/clock"

// BeadFraction returns a bead's clamped fractional progress along a wire (0..1)
// at nowTick, given the tick it was placed at (placementTick) and the ticks the
// whole crossing takes (crossTicks = steps * DwellTicksPerBead). Progress never
// runs past 1 even once nowTick has overshot the crossing deadline (a bead may
// sit finished-but-undelivered at the FIFO head for a cycle or more while it
// waits on a full out-channel) and never runs below 0 (a bead read before its own
// placement tick, which ReviseInFlightGeometry's callers can hit). Returns 0 when
// crossTicks<=0 — there is no crossing duration to measure a fraction against.
func BeadFraction(nowTick, placementTick, crossTicks float64) float64 {
	if crossTicks <= 0 {
		return 0
	}
	target := nowTick
	if nowTick >= placementTick+crossTicks {
		target = placementTick + crossTicks
	}
	t := (target - placementTick) / crossTicks
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return t
}

// SimLatencyMs is the REPORTED diagnostic latency for a bead crossing an edge of
// the given bead-step count, derived from the uniform per-step dwell — not an
// independently measured value (out_port.go's flushSendEvent, the SendWire trace's
// SimLatencyMs field).
func SimLatencyMs(steps int) float64 {
	return float64(steps) * DwellTicksPerBead * clock.MsPerTick
}
