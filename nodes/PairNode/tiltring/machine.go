package tiltring

// machine.go — THE TILT STATE MACHINE. One machine, one rule, one file.
//
// THE COUNT is the one measurement: how many ring slots lie between the arrival and the NEARER END
// of a node's tilt line. A node draws two ends, t and t + half, so there are two counts and they
// differ by a half turn — exactly one is under a half turn, and that one is the acute angle at its
// own end. So the count lives on a ring of a HALF TURN. Its VALUE says what the pair is at, and the
// end it was taken at is the end an update moves.
//
// A mode is the counts it STOPS at, and that row is the only thing that differs between the three.
// The rule that turns toward them is written ONCE.
//
//	setting        stops at { every one }   which machine to run is still being decided
//	perpendicular  stops at { 0 }           the arrival lies ON this node's tilt line
//	parallel       stops at { quarter }     the arrival lies a quarter turn off it
//
// ONE ARRIVAL ANSWERS TWO YES-OR-NO QUESTIONS AND NOTHING MORE: is this count a stopping one — if
// so the node settles and sends nothing — and if not, is the stop nearer going up the count-ring or
// down it. Then it turns ONE slot. No walk is planned and no distance is kept; the next arrival
// re-derives everything from scratch.
//
// See docs/pair-node/rules/audit.html and docs/pair-node/math/arith.html for the derivation this
// file states as code; PairNode's own SPEC.md and machine.go (the two remaining Node methods,
// machineForGap/adoptMachine) carry the parts of this story that are about what a NODE does with
// the machine rather than what the machine itself is.

import (
	"strconv"

	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

// Machine is THE machine — one instance per node, not one object per mode. What it carries is
// which mode it is in, and that is the whole of its state; the stopping counts are read from the
// mode (Stopping) rather than stored, so there is nothing per-mode to construct, share, or get out
// of step with the mode itself.
//
// It is a VALUE, and its ZERO VALUE IS THE SETTING MODE — TiltMachineNone is the zero of
// tiltvector.TiltMachine. A node built as a literal is therefore already in the mode a node starts
// in, with no constructor to call and no nil to test for. There is no second spelling of any mode,
// so two machines are equal exactly when they are in the same mode.
type Machine struct {
	Mode tiltvector.TiltMachine
}

// NearerEndCount is THE measurement: how far the arrival is from the nearer END OF THIS STATE'S
// TILT LINE.
//
// A node draws two ends, t and t + h, so there are two counts to the arrival and they differ by a
// half turn — exactly one is under h, and that one is already the acute angle at its own end, with
// nothing left to compute and nothing to fold. So c is a count on a ring of a HALF TURN, not of the
// whole lattice.
//
// Which end it came from is returned with it because an update moves THAT end.
func (s *State) NearerEndCount(arrival *State) (c int32, atBottom bool) {
	u := ((s.Idx-arrival.Idx)%s.R.Points + s.R.Points) % s.R.Points
	if u < s.R.HalfTurn {
		return u, false
	}
	// The bottom's own count, (b - a) mod points, and NOT u with a half turn folded out of it: b
	// is a state the caller already has, and the two agree because b = t + h.
	return ((s.Opposite.Idx-arrival.Idx)%s.R.Points + s.R.Points) % s.R.Points, true
}

// StoppingCount is the per-mode data, and the ONLY per-mode thing there is: which counts on the
// half-turn ring that mode STOPS at. A new mode is a new entry in stoppingCounts — not a new type,
// not a new file, and not a branch anywhere in the rule below.
//
// A ROW IS ONE COUNT, NOT A LIST: measured from the nearer end there is nowhere for a second entry
// to come from. `Anywhere` is the one thing a row can say instead of naming a count.
type StoppingCount struct {
	// Anywhere: this mode stops at EVERY count, so it is at rest wherever it stands.
	Anywhere bool
	// At: the count it stops at, read off the ring. Meaningless when Anywhere is set.
	At func(r *Ring) int32
}

var stoppingCounts = map[tiltvector.TiltMachine]StoppingCount{
	// Setting: which machine to run is still being decided, so NOTHING IS OUT OF PLACE YET —
	// every count is a stopping one. That is what makes a node being set up hold still by the
	// ordinary rule instead of by an exemption from it: it is settled wherever it stands, so
	// Step is never reached.
	tiltvector.TiltMachineNone: {Anywhere: true},
	// The arrival lies ON this node's tilt line — either end of it, which from the nearer end
	// is one count and not two.
	tiltvector.TiltMachinePerpendicular: {At: func(r *Ring) int32 { return 0 }},
	// The arrival lies a quarter turn off it, on the normal line.
	tiltvector.TiltMachineParallel: {At: func(r *Ring) int32 { return r.QuarterTurn }},
}

// Stopping is this machine's row — the only per-mode data there is.
func (m Machine) Stopping() StoppingCount { return stoppingCounts[m.Mode] }

// Settled is the FIRST of the two questions an arrival asks: is this count the one this mode stops
// at? A COMPARISON, and nothing else.
func (m Machine) Settled(from, arrival *State) bool {
	stop := m.Stopping()
	if stop.Anywhere {
		return true
	}
	c, _ := from.NearerEndCount(arrival)
	return c == stop.At(from.R)
}

// Step answers the SECOND question — up or down — and turns ONE slot that way. It is a link
// either way, so the turn cannot leave the ring.
//
// ONE SUBTRACTION AND ONE COMPARISON. How far the stop is going UP is (stop − c) mod h; the way
// down is the rest of the circle, h minus that, so asking which is shorter is asking whether the
// upward count is at most a quarter turn. There is no downward count to work out and nothing to
// take a minimum over, and a tie — exactly a quarter each way — goes up by the ≤.
//
// IT RETURNS THE END THAT MOVED, AND WHICH END THAT IS. The count was taken at the nearer end, so
// the slot it gains is that end's slot. Turning either end moves both, so the caller writes the
// other from this one's opposite; naming the driven end is what lets it be written without a
// correction term for the half where the bottom was nearer.
//
// A mode that stops anywhere never reaches here — Settled above is already true — so its row
// naming no count costs this nothing.
func (m Machine) Step(from, arrival *State) (moved *State, atBottom bool) {
	c, atBottom := from.NearerEndCount(arrival)
	end := from
	if atBottom {
		end = from.Opposite
	}
	h := from.R.HalfTurn
	if ((m.Stopping().At(from.R)-c)%h+h)%h <= from.R.QuarterTurn {
		return end.Next, atBottom
	}
	return end.Prev, atBottom
}

// Choice is this machine's pair-wide name, so the end that chose can tell the other one
// (tiltvector.TiltMachine, carried on every vector message). It is the mode itself: the name this
// package runs on and the name the two ends say to each other are ONE value, so there is no
// mapping to keep honest in either direction.
func (m Machine) Choice() tiltvector.TiltMachine { return m.Mode }

// String names the mode for the diagnostic row — the modes have to be distinguishable there, since
// a log that printed them alike is what once hid two of them being one state.
func (m Machine) String() string {
	if m.Mode == tiltvector.TiltMachineNone {
		return "setting"
	}
	return m.Mode.String()
}

// Setting is the mode a node starts in and returns to on a reset. It is spelled out here so
// readers that ask "is this node still being set up?" can say so by name rather than by comparing
// against a bare zero value.
var Setting = Machine{Mode: tiltvector.TiltMachineNone}

// MachineFor is the machine a node runs for a pair-wide choice. There is no mapping table and no
// per-mode object to look up: the choice IS the mode, so this is the machine holding it. A choice
// this package does not recognise has no stopping counts, and running it would panic on the nil
// lookup rather than silently behaving like some other mode — so it is refused here, where the
// name of the broken invariant can still be said.
func MachineFor(choice tiltvector.TiltMachine) Machine {
	if _, known := stoppingCounts[choice]; !known {
		// The numeric value, not choice.String(): that falls back to "none" for anything it
		// does not recognise, so it would name the wrong mode in exactly this message.
		panic("tiltring: no stopping counts for tilt machine " + strconv.Itoa(int(choice)) +
			" — every mode must name the counts it stops at (machine.go stoppingCounts)")
	}
	return Machine{Mode: choice}
}
