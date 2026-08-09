// drive_item.go — ONE JOB: what a caller is TOLD about one placement. The
// DriveOutcome enum that keeps "placed", "chan-mode sent", "transiently full" and
// "structurally failed" four distinct answers, the DriveItem handle that reports
// which one happened, and Out.PlaceDrivenAt, the entry point that performs a
// placement and answers in those terms. The placement mechanics themselves are
// out_port.go; placing on a whole set of Outs at once is broadcast.go.

package wire

// DriveOutcome distinguishes the outcomes a drive placement can have.
// Collapsing them onto a single bool (the pre-fix shape) made "chan mode, sent
// fine, nothing more to drive" indistinguishable from "placement failed /
// wire torn down" — callers that stopped their loop on !Live() then stopped
// on every chan-mode send too. Keeping them explicit makes that conflation
// unrepresentable.
//
// DriveBufferFull is its own constant, split out from DriveFailed, for the
// same reason: a paced wire's inCh being momentarily full (PacedWire.
// SendBufferFull) is TRANSIENT — the wire's own goroutine drains inCh every
// cycle — and must never be treated as "stop, the wire is gone". Only a nil
// Out (no destination at all, a structural condition) is DriveFailed. Naming
// them apart means a caller who writes `.Failed()` gets ONLY the terminal
// case by construction; it cannot accidentally also catch buffer-full the way
// the pre-fix single bool did.
type DriveOutcome uint8

const (
	// DrivePlaced: a real bead was placed on a paced wire; delivery is driven
	// by subsequent StepOnce/StepOnceAt calls.
	DrivePlaced DriveOutcome = iota
	// DriveSentChan: chan mode (tests) — the value was sent (or dropped by a
	// full non-blocking select) on the raw channel. Nothing to drive further.
	DriveSentChan
	// DriveBufferFull: the paced wire's inCh was momentarily full
	// (PacedWire.SendBufferFull). TRANSIENT, NOT TERMINAL — a caller driving a
	// continuous-placement loop must NOT stop on this; skip this cycle's
	// placement and retry next cycle. A breadcrumb was already emitted by
	// PacedWire.Send.
	DriveBufferFull
	// DriveFailed: nil Out — there is no destination at all. Structural and
	// terminal; the caller should stop.
	DriveFailed
)

// DriveItem is an exported handle to one placed bead. Delivery is timed by the
// wire's own goroutine, not by the caller — this type reports which of the
// DriveOutcomes occurred.
type DriveItem struct {
	outcome DriveOutcome
}

// Live reports whether this DriveItem carries a bead actually placed on a
// paced wire (i.e. PlaceDriven succeeded in paced-wire mode) — outcome ==
// DrivePlaced. False for a nil Out, chan mode, a momentary buffer-full, or a
// failed placement. Callers that need ONLY "did this become a real, time-able
// in-flight bead" (e.g. Time's processing-window length) check
// this; callers that need "should I stop, the wire is gone" must check
// Failed() instead — Live() alone cannot distinguish chan-mode success,
// buffer-full, or true failure.
func (di DriveItem) Live() bool {
	return di.outcome == DrivePlaced
}

// Failed reports whether the placement failed for a STRUCTURAL, TERMINAL
// reason: a nil Out (no destination at all). It deliberately does NOT report
// true for DriveBufferFull — a momentarily-full paced-wire buffer is
// transient and self-clears as the wire's own goroutine drains it; treating
// it as terminal would silently and permanently kill a source's drive
// goroutine on ordinary transient load. Callers driving a continuous-
// placement loop should stop on Failed(), not on !Live() — a chan-mode
// successful send or a buffer-full retry are also !Live() but must not stop
// the loop. See BufferFull() for the transient case.
func (di DriveItem) Failed() bool {
	return di.outcome == DriveFailed
}

// BufferFull reports whether this placement did not go through because the
// paced wire's inCh was momentarily full. This is the DISTINCT, NON-TERMINAL
// counterpart to Failed(): a caller must handle this case (typically: skip
// this cycle, keep looping) rather than let it fall through to a generic
// failure branch, which is exactly the bug this type split fixes.
func (di DriveItem) BufferFull() bool {
	return di.outcome == DriveBufferFull
}

// PlaceDrivenAt places one bead on this Out WITHOUT spawning a walker, emits
// the SendWire trace, and returns a DriveItem reporting the outcome. tick is
// the CALLING goroutine's own clock reading, read once by the caller and
// stamped as this bead's placementTick — the wire itself no longer decides
// when a bead started (placeRequest's doc comment). Delivery timing (the
// bead's later position-advance) is still done by the wire's driver (the
// source node's mover — node_mover.go), not the caller. In chan mode (tests)
// it sends immediately on the raw channel and returns DriveSentChan, so unit
// tests keep their synchronous chan semantics (tick is unused in this path). A
// nil Out returns DriveFailed. A momentarily-full paced wire returns
// DriveBufferFull, never DriveFailed — see the DriveOutcome doc comment.
func (o *Out) PlaceDrivenAt(v int, tick int64) DriveItem {
	if o == nil {
		return DriveItem{outcome: DriveFailed}
	}
	if o.pw != nil {
		switch o.placeDrivenNoWalker(v, tick) {
		case SendPlaced:
			return DriveItem{outcome: DrivePlaced}
		default: // SendBufferFull
			return DriveItem{outcome: DriveBufferFull}
		}
	}
	// chan mode (tests, or a production dead-end unwired Out): no drive needed, send
	// now and return DriveSentChan. flushSendEvent no-ops when stream is unset (both
	// cases: no builder-injected getter).
	if o.ch != nil {
		select {
		case o.ch <- v:
			o.flushSendEvent(v, 0)
		default:
		}
		return DriveItem{outcome: DriveSentChan}
	}
	return DriveItem{outcome: DriveSentChan}
}
