package Node1

// machine.go — THE TILT STATE MACHINE. One machine, one rule, one file.
//
// THE COUNT is the one measurement: how many ring slots lie between the arrival and the NEARER END
// of this node's tilt line. A node draws two ends, t and t + half, so there are two counts and they
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
// THE SECOND QUESTION IS ASKED FROM WHERE THE NODE IS, not from the two places it could go. It
// used to fold the arrival to an angle length, hold that against a list, and then redo that whole
// measurement at BOTH NEIGHBOURS so the two results could be compared — a machine choosing between
// its next two positions without ever working out where it was heading. What docs/pair-node/arith.html
// states, and what this file now computes, is one count compared to one number.
//
// Setting is a mode like the other two and not an absence of one: while a node is still being set
// up nothing is out of place, so every count is a stopping one and the node holds still by the
// ordinary rule rather than by an exemption from it.
//
// WHY THIS IS ONE FILE NOW, HAVING DELIBERATELY BEEN TWO.
//
// perpendicular.go and parallel.go were split apart because three attempts at changing one rule
// changed the other as a side effect — every time through something that served both: a fromRest
// function PARAMETERIZED BY WHICH STATE YOU WERE IN, a step that took the state as an ARGUMENT,
// and a version where a node holding neither compared its distance to parallel against its
// distance to perpendicular to pick a target. Two files had nowhere to put that coupling, and
// that worked: both rules were finished without either disturbing the other.
//
// The audit that followed (docs/pair-node/audit.html) measured what the two had become. Across
// 103 lines: step was character-for-character identical; settled was each machine restating its
// own fromRest being zero; String and choice were mirrored boilerplate; and the two fromRest values are EXACT
// COMPLEMENTS -- perpendicular fromRest = quarter − parallel fromRest, everywhere on the ring, proved on
// both sides of the fold. One rule, run in opposite directions on one number. The genuinely
// distinct behaviour was one line of data — then a list of resting lengths, now the counts a mode
// stops at.
//
// So the fold here is NOT the fourth attempt at the three that failed. Those three parameterized
// the RULE — a branch inside fromRest, a mode passed to step — which is what let a change meant for
// one mode land inside the other. Nothing here branches on which mode is running: walk is a
// distance to a set of numbers and cannot ask who it is working for, step reads only which of its
// two answers is smaller, and settled is that minimum being zero. A mode contributes its stopping
// counts and NOTHING ELSE, so there is no code belonging to one mode for a change to the other to
// reach into. The coupling the split defended against has no place left to live.
//
// THAT SURVIVED THE MOVE TO THE PAGE'S ARITHMETIC, which is not obvious: arith.html prints the
// update as two four-line blocks, one per arrangement, and they LOOK like two rules. They are one
// sentence said twice — walk the count the short way to your own stop — and stating it that way is
// what keeps the mode out of the rule. Written as the page prints it, this file would have a
// switch in it and the paragraph above would be false.
//
// WHAT ARRIVES is the partner's coplanar NORMAL, already a quarter turn off the partner's own
// tilt. That is why the stopping counts read the way they do: "the tilts are a quarter turn apart"
// reaches a node as the arrival ON its own tilt line, either end, which from the nearer end is the
// single count 0 — and "the tilts lie on one line" reaches it a quarter turn off that line.

import (
	"strconv"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// tiltMachine is THE machine — one instance per node, not one object per mode. What it carries is
// which mode it is in, and that is the whole of its state; the stopping counts are read from the mode
// (stoppingCounts) rather than stored, so there is nothing per-mode to construct, share, or get out of
// step with the mode itself.
//
// It is a VALUE, and its ZERO VALUE IS THE SETTING MODE — TiltMachineNone is the zero of
// Wiring.TiltMachine. A Node built as a literal is therefore already in the mode a node starts in,
// with no constructor to call and no nil to test for. There is no second spelling of any mode, so
// two machines are equal exactly when they are in the same mode.
type tiltMachine struct {
	mode Wiring.TiltMachine
}

// nearerEndCount is THE measurement, and it is the one docs/pair-node/arith.html makes: how far the
// arrival is from the nearer END OF THIS NODE'S TILT LINE.
//
// A node draws two ends, t and t + h, so there are two counts to the arrival and they differ by a
// half turn — exactly one is under h, and that one is already the acute angle at its own end, with
// nothing left to compute and nothing to fold. So c is a count on a ring of a HALF TURN, not of the
// whole lattice.
//
// Which end it came from is returned with it because the update moves THAT end.
func (s *tiltState) nearerEndCount(arrival *tiltState) (c int32, atBottom bool) {
	u := ((s.idx-arrival.idx)%s.ring.points + s.ring.points) % s.ring.points
	if u < s.ring.halfTurn {
		return u, false
	}
	// The bottom's own count, (b - a) mod points, and NOT u with a half turn folded out of it:
	// b is a state this node already has, and the two agree because b = t + h.
	return ((s.opposite.idx-arrival.idx)%s.ring.points + s.ring.points) % s.ring.points, true
}

// stoppingCounts is the per-mode data, and the ONLY per-mode thing there is: which counts on the
// half-turn ring that mode STOPS at. A new mode is a new entry here — not a new type, not a new
// file, and not a branch anywhere in the rule below.
//
// This was a list of RESTING LENGTHS, measured from the top, and it needed two entries wherever a
// mode's two stopping tilts fell on opposite ends of the line. Measured from whichever end is
// NEARER, perpendicular's { 0, half } is a single 0 — the two entries were one fact counted twice.
//
// Each entry takes the ring: a quarter turn is 12 points on a 48-point ring and 6 on a 24-point
// one, and a node can change lattice underneath a running machine (Node.adoptLattice).
var stoppingCounts = map[Wiring.TiltMachine]func(r *ring) []int32{
	// Setting: which machine to run is still being decided, so NOTHING IS OUT OF PLACE YET —
	// every count is a stopping one. That is what makes a node being set up hold still by the
	// ordinary rule instead of by an exemption from it: fromRest is zero wherever it stands, so
	// it is always settled, so step is never reached.
	Wiring.TiltMachineNone: func(r *ring) []int32 {
		// c is a count on the half-turn ring, so this is every count there is.
		all := make([]int32, 0, r.halfTurn)
		for c := int32(0); c < r.halfTurn; c++ {
			all = append(all, c)
		}
		return all
	},
	// The arrival lies ON this node's tilt line — either end of it, which from the nearer end
	// is one count and not two.
	Wiring.TiltMachinePerpendicular: func(r *ring) []int32 { return []int32{0} },
	// The arrival lies a quarter turn off it, on the normal line.
	Wiring.TiltMachineParallel: func(r *ring) []int32 { return []int32{r.quarterTurn} },
}

// stopping is this machine's stopping counts on the given ring — its mode's row in
// stoppingCounts, and the only per-mode data there is.
func (m tiltMachine) stopping(r *ring) []int32 { return stoppingCounts[m.mode](r) }

// walk is how far c is from this mode's nearest stopping count, each way round the half-turn ring.
// Both are non-negative counts of slots; the smaller one is the direction, and their minimum is
// fromRest.
//
// It is where the mode enters the arithmetic, and it enters as DATA — a row of counts — so nothing
// here learns which mode it is computing for.
func (m tiltMachine) walk(from, arrival *tiltState) (up, down int32) {
	h := from.ring.halfTurn
	c, _ := from.nearerEndCount(arrival)
	up, down = h, h
	for _, stop := range m.stopping(from.ring) {
		if d := ((stop-c)%h + h) % h; d < up {
			up = d
		}
		if d := ((c-stop)%h + h) % h; d < down {
			down = d
		}
	}
	return up, down
}

// fromRest is how far this arrival's count is from a stopping one — zero when it already IS one.
// NOTHING KEEPS THIS. It is computed inside one arrival and discarded. It is not a countdown and no
// walk is planned from it: one arrival moves the tilt at most one slot.
func (m tiltMachine) fromRest(from, arrival *tiltState) int32 {
	return min(m.walk(from, arrival))
}

// settled is the FIRST of the two questions an arrival asks: is this count one this mode stops at?
// It is fromRest being zero and nothing else — asking it a second, independent way is how the two
// machines each grew a predicate that restated their own measurement.
func (m tiltMachine) settled(from, arrival *tiltState) bool { return m.fromRest(from, arrival) == 0 }

// step answers the SECOND question — up or down — and turns ONE slot that way. It is a link either
// way, so the turn cannot leave the ring.
//
// THE COMPARISON IS BETWEEN THE TWO WAYS ROUND FROM WHERE THIS NODE IS, not between two places it
// might go: c walks the short way to its stop, and a tie goes UP. What the page prints as two
// four-line blocks — parallel stopping at a quarter turn, perpendicular at 0 — is this one
// sentence twice, which is why neither block appears here.
//
// IT RETURNS THE END THAT MOVED, AND WHICH END THAT IS. The count was taken at the nearer end, so
// the slot it gains is that end's slot — the page's two halves each measure an end and then move
// the one they measured. Turning either end moves both, so the caller writes the other from this
// one's opposite; naming the driven end is what lets it be written without a correction term for
// the half where the bottom was nearer.
func (m tiltMachine) step(from, arrival *tiltState) (moved *tiltState, atBottom bool) {
	_, atBottom = from.nearerEndCount(arrival)
	end := from
	if atBottom {
		end = from.opposite
	}
	if up, down := m.walk(from, arrival); up <= down {
		return end.next, atBottom
	}
	return end.prev, atBottom
}

// choice is this machine's pair-wide name, so the end that chose can tell the other one
// (Wiring.TiltMachine, carried on every vector message). It is the mode itself: the name this
// package runs on and the name the two ends say to each other are ONE value, so there is no
// mapping to keep honest in either direction.
func (m tiltMachine) choice() Wiring.TiltMachine { return m.mode }

// String names the mode for the diagnostic row — the modes have to be distinguishable there, since
// a log that printed them alike is what once hid two of them being one state.
func (m tiltMachine) String() string {
	if m.mode == Wiring.TiltMachineNone {
		return "setting"
	}
	return m.mode.String()
}

// setting is the mode a node starts in and returns to on a reset. It is spelled out here so the
// readers that ask "is this node still being set up?" can say so by name rather than by comparing
// against a bare zero value.
var setting = tiltMachine{mode: Wiring.TiltMachineNone}

// machineFor is the machine a node runs for a pair-wide choice. There is no mapping table and no
// per-mode object to look up: the choice IS the mode, so this is the machine holding it. A choice
// this package does not recognise has no stopping counts, and running it would panic on the nil lookup
// rather than silently behaving like some other mode — so it is refused here, where the name of
// the broken invariant can still be said.
func machineFor(choice Wiring.TiltMachine) tiltMachine {
	if _, known := stoppingCounts[choice]; !known {
		// The numeric value, not choice.String(): that falls back to "none" for anything it
		// does not recognise, so it would name the wrong mode in exactly this message.
		panic("Node1: no stopping counts for tilt machine " + strconv.Itoa(int(choice)) +
			" — every mode must name the counts it stops at (machine.go stoppingCounts)")
	}
	return tiltMachine{mode: choice}
}
