package Node1

// machine.go — THE TILT STATE MACHINE. One machine, one rule, one file.
//
// ANGLE LENGTH is the one measurement: how many ring slots lie between this node's own tilt and
// the direction that arrived, counted the short way round, so never more than a half turn. Its
// VALUE says what the pair is at.
//
// A mode is a list of the angle lengths it RESTS at, and that list is the only thing that differs
// between the three. The rule that turns toward them is written ONCE.
//
//	setting        rests at { every one }   which machine to run is still being decided
//	perpendicular  rests at { 0, half }     the two tilts a quarter turn apart
//	parallel       rests at { quarter }     the two tilts on the SAME LINE, either way round
//
// ONE ARRIVAL ANSWERS TWO YES-OR-NO QUESTIONS AND NOTHING MORE: is this angle length a resting
// one — if so the node settles and sends nothing — and if not, is the neighbour above closer to a
// resting length than the one below. Then it turns ONE slot. No walk is planned and no distance is
// kept; the next arrival re-derives everything from scratch.
//
// Setting is a mode like the other two and not an absence of one: while a node is still being set
// up nothing is out of place, so every angle length is a resting state and the node holds still by
// the ordinary rule rather than by an exemption from it.
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
// distinct behaviour was the resting-length list: one line of data.
//
// So the fold here is NOT the fourth attempt at the three that failed. Those three parameterized
// the RULE — a branch inside fromRest, a mode passed to step — which is what let a change meant for
// one mode land inside the other. Nothing here branches on which mode is running: fromRest is a
// minimum over a set of numbers and cannot ask who it is working for, step never sees a mode at
// all, and settled is fromRest being zero. A mode contributes its resting-length list and NOTHING ELSE, so there
// is no code belonging to one mode for a change to the other to reach into. The coupling the
// split defended against has no place left to live.
//
// WHAT ARRIVES is the partner's coplanar NORMAL, already a quarter turn off the partner's own
// tilt. That is why the resting-length lists read the way they do: "the tilts are a quarter turn apart"
// reaches a node as the arrival on its own TOP or exactly opposite it (angle length 0 or half),
// and "the tilts lie on one line" reaches it as a quarter turn off its top.

import (
	"strconv"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// tiltMachine is THE machine — one instance per node, not one object per mode. What it carries is
// which mode it is in, and that is the whole of its state; the resting-length list is read from the mode
// (restingLengths) rather than stored, so there is nothing per-mode to construct, share, or get out of
// step with the mode itself.
//
// It is a VALUE, and its ZERO VALUE IS THE SETTING MODE — TiltMachineNone is the zero of
// Wiring.TiltMachine. A Node built as a literal is therefore already in the mode a node starts in,
// with no constructor to call and no nil to test for. There is no second spelling of any mode, so
// two machines are equal exactly when they are in the same mode.
type tiltMachine struct {
	mode Wiring.TiltMachine
}

// restingLengths is the per-mode data, and the ONLY per-mode thing there is: which angle lengths that mode
// calls a resting state. A new mode is a new entry here — not a new type, not a new file, and not a
// branch anywhere in the rule below.
//
// A resting length is a count of slots on the lattice, so each entry takes the ring: a quarter turn is 12 points on
// a 48-point ring and 6 on a 24-point one, and a node can change lattice underneath a running
// machine (Node.adoptLattice).
//
// perpendicular and parallel are not independent: quarter is the midpoint of 0 and half, which is
// why their two fromRest values sum to a constant quarter and why each one's resting length is the other's farthest
// point (docs/pair-node/audit.html).
var restingLengths = map[Wiring.TiltMachine]func(r *ring) []int32{
	// Setting: which machine to run is still being decided, so NOTHING IS OUT OF PLACE YET —
	// every angle length is a resting state. That is what makes a node being set up hold still by
	// the ordinary rule instead of by an exemption from it: fromRest is zero wherever it stands, so
	// it is always settled, so step is never reached.
	Wiring.TiltMachineNone: func(r *ring) []int32 {
		// angle length folds into [0, half], so this is every angle length there is.
		all := make([]int32, 0, r.halfTurn+1)
		for sep := int32(0); sep <= r.halfTurn; sep++ {
			all = append(all, sep)
		}
		return all
	},
	// The two tilts a quarter turn apart, which reaches a node as the arrival on its own TOP or
	// exactly opposite it.
	Wiring.TiltMachinePerpendicular: func(r *ring) []int32 {
		return []int32{0, r.halfTurn}
	},
	// The two tilts on the SAME LINE, either way round, which reaches a node as a quarter turn
	// off its own top.
	Wiring.TiltMachineParallel: func(r *ring) []int32 {
		return []int32{r.quarterTurn}
	},
}

// resting is this machine's resting angle lengths on the given ring — its mode's row in
// restingLengths, and the only per-mode data there is.
func (m tiltMachine) resting(r *ring) []int32 { return restingLengths[m.mode](r) }

// fromRest is how far this arrival's angle length is from a resting one — zero when it already IS
// one. NOTHING KEEPS THIS. It is computed inside one arrival and answers two questions there — is
// this settled, and which of the two neighbours is closer — after which it is discarded. It is not
// a countdown and no walk is planned from it: one arrival moves the tilt at most one slot.
//
// It is also where the mode enters the arithmetic, and it enters as DATA — a list of lengths — so
// this function never learns which mode it is computing for.
func (m tiltMachine) fromRest(from, arrival *tiltState) int32 {
	length := from.angleLength(arrival)
	resting := m.resting(from.ring)
	nearest := gapBetween(length, resting[0])
	for _, r := range resting[1:] {
		if d := gapBetween(length, r); d < nearest {
			nearest = d
		}
	}
	return nearest
}

// settled is the FIRST of the two questions an arrival asks: is this angle length one this mode
// rests at? It is fromRest being zero and nothing else — asking it a second, independent way is
// how the two machines each grew a predicate that restated their own measurement.
func (m tiltMachine) settled(from, arrival *tiltState) bool { return m.fromRest(from, arrival) == 0 }

// step answers the SECOND question — up or down — and turns ONE slot that way. It is a link either
// way, so the turn cannot leave the ring. The comparison is the whole of it: whichever neighbour is
// closer to a resting length wins, and a tie goes UP. Where a mode rests at more than one length
// this lands on whichever is reached first; no choosing between them happens anywhere.
func (m tiltMachine) step(from, arrival *tiltState) *tiltState {
	if m.fromRest(from.next, arrival) <= m.fromRest(from.prev, arrival) {
		return from.next
	}
	return from.prev
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
// this package does not recognise has no resting-length list, and running it would panic on the nil lookup
// rather than silently behaving like some other mode — so it is refused here, where the name of
// the broken invariant can still be said.
func machineFor(choice Wiring.TiltMachine) tiltMachine {
	if _, known := restingLengths[choice]; !known {
		// The numeric value, not choice.String(): that falls back to "none" for anything it
		// does not recognise, so it would name the wrong mode in exactly this message.
		panic("Node1: no resting-length list for tilt machine " + strconv.Itoa(int(choice)) +
			" — every mode must name the angle lengths it rests at (machine.go restingLengths)")
	}
	return tiltMachine{mode: choice}
}
