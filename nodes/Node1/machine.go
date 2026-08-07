package Node1

// machine.go — THE TILT STATE MACHINE. One machine, one rule, one file.
//
// A pair node walks its tilt until the arrival sits at a separation this machine calls HOME.
// Which separations those are is the ONLY thing that differs between the modes, so it is the
// only thing written per mode: a home set, one line of data each, in the table at the bottom of
// this file. The rule that walks toward them is written ONCE.
//
//	perpendicular  home = { 0, half }     the two tilts a quarter turn apart
//	parallel       home = { quarter }     the two tilts on the SAME LINE, either way round
//
// WHY THIS IS ONE FILE NOW, HAVING DELIBERATELY BEEN TWO.
//
// perpendicular.go and parallel.go were split apart because three attempts at changing one rule
// changed the other as a side effect — every time through something that served both: a miss
// function PARAMETERIZED BY WHICH STATE YOU WERE IN, a step that took the state as an ARGUMENT,
// and a version where a node holding neither compared its distance to parallel against its
// distance to perpendicular to pick a target. Two files had nowhere to put that coupling, and
// that worked: both rules were finished without either disturbing the other.
//
// The audit that followed (docs/pair-node/audit.html) measured what the two had become. Across
// 103 lines: step was character-for-character identical; halted was each machine restating its
// own miss being zero; String and choice were mirrored boilerplate; and the two misses are EXACT
// COMPLEMENTS -- perpendicular miss = quarter − parallel miss, everywhere on the ring, proved on
// both sides of the fold. One rule, run in opposite directions on one number. The genuinely
// distinct behaviour was the home set: one line of data.
//
// So the fold here is NOT the fourth attempt at the three that failed. Those three parameterized
// the RULE — a branch inside miss, a mode passed to step — which is what let a change meant for
// one mode land inside the other. Nothing here branches on which mode is running: miss is a
// minimum over a set of numbers and cannot ask who it is working for, step never sees a mode at
// all, and halted is miss being zero. A mode contributes its home set and NOTHING ELSE, so there
// is no code belonging to one mode for a change to the other to reach into. The coupling the
// split defended against has no place left to live.
//
// WHAT ARRIVES is the partner's coplanar NORMAL, already a quarter turn off the partner's own
// tilt. That is why the home sets read the way they do: "the tilts are a quarter turn apart"
// reaches a node as the arrival on its own TOP or exactly opposite it (separation 0 or half),
// and "the tilts lie on one line" reaches it as a quarter turn off its top.

import "github.com/dtauraso/wirefold/nodes/Wiring"

// tiltMachine is the one machine. A mode is an instance of it carrying a different home set;
// there is no second implementation and no interface, because there is no second rule.
type tiltMachine struct {
	name string
	pick Wiring.TiltMachine
	// homes are the separations that ARE this mode's resting state, measured on the ring the
	// node is holding. It takes the ring because a home is a position on the lattice — a
	// quarter turn is 12 points on a 48-point ring and 6 on a 24-point one — and a node can
	// change lattice underneath a running machine (Node.adoptLattice).
	homes func(r *ring) []int32
}

// miss is how far this arrival is from the NEAREST of this machine's resting states — zero when
// it is one of them. This is the whole per-mode difference, and it is a minimum over data: the
// function never learns which mode it is computing for.
func (m *tiltMachine) miss(from, arrival *tiltState) int32 {
	sep := from.separation(arrival)
	homes := m.homes(from.ring)
	nearest := abs32(sep - homes[0])
	for _, h := range homes[1:] {
		if d := abs32(sep - h); d < nearest {
			nearest = d
		}
	}
	return nearest
}

// halted reports whether this arrival IS one of this machine's resting states. A machine is home
// exactly when it has nothing left to close, which is what miss already measures — asking it a
// second way is how the two machines each grew a predicate that restated their own miss.
func (m *tiltMachine) halted(from, arrival *tiltState) bool { return m.miss(from, arrival) == 0 }

// step is the single move — next or prev, a link either way, so it cannot leave the ring — that
// leaves the node closer to its halt. Where a mode has more than one resting state this closes on
// whichever is nearer, because miss already reports the nearest.
func (m *tiltMachine) step(from, arrival *tiltState) *tiltState {
	if m.miss(from.next, arrival) <= m.miss(from.prev, arrival) {
		return from.next
	}
	return from.prev
}

// choice is this machine's pair-wide name, so the end that chose can tell the other one
// (Wiring.TiltMachine, carried on every vector message).
func (m *tiltMachine) choice() Wiring.TiltMachine { return m.pick }

// String names the mode for the diagnostic row — the modes have to be distinguishable there,
// since a log that printed them alike is what once hid two of them being one state.
func (m *tiltMachine) String() string { return m.name }

// THE MODES. Each is its home set and its two names, and a new mode is a new row here — not a
// new file, and not a new branch anywhere above.
//
// The two sets are not independent: quarter is the midpoint of 0 and half, which is why the two
// misses sum to a constant quarter and why each mode's home is the other's farthest point.
var (
	perpendicularMachine = &tiltMachine{
		name: "perpendicular",
		pick: Wiring.TiltMachinePerpendicular,
		homes: func(r *ring) []int32 {
			return []int32{0, r.halfTurn}
		},
	}

	parallelMachine = &tiltMachine{
		name: "parallel",
		pick: Wiring.TiltMachineParallel,
		homes: func(r *ring) []int32 {
			return []int32{r.quarterTurn}
		},
	}
)

// machineFor maps the pair-wide name onto the mode this package runs for it. The naming lives in
// Wiring so both ends can say it to each other; the home sets live here, which is the only place
// that knows what any of them means on the ring.
//
// TiltMachineNone maps to no machine, which is what a reset restores and what an ordinary vector
// message carries.
func machineFor(choice Wiring.TiltMachine) *tiltMachine {
	switch choice {
	case Wiring.TiltMachinePerpendicular:
		return perpendicularMachine
	case Wiring.TiltMachineParallel:
		return parallelMachine
	}
	return nil
}
