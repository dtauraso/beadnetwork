package Node1

// machine.go — THE TILT STATE MACHINE. One machine, one rule, one file.
//
// A pair node walks its tilt until the arrival sits at a separation this machine calls HOME.
// Which separations those are is the ONLY thing that differs between the modes, so it is the
// only thing written per mode: a home set, one line of data each, in the table at the bottom of
// this file. The rule that walks toward them is written ONCE.
//
//	setting        home = { every one }   which machine to run is still being decided
//	perpendicular  home = { 0, half }     the two tilts a quarter turn apart
//	parallel       home = { quarter }     the two tilts on the SAME LINE, either way round
//
// Setting is a mode like the other two and not an absence of one: while a node is still being set
// up nothing is out of place, so every separation is a resting state and the node holds still by
// the ordinary rule rather than by an exemption from it.
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

import (
	"strconv"

	"github.com/dtauraso/wirefold/nodes/Wiring"
)

// tiltMachine is THE machine — one instance per node, not one object per mode. What it carries is
// which mode it is in, and that is the whole of its state; the home set is read from the mode
// (homeSets) rather than stored, so there is nothing per-mode to construct, share, or get out of
// step with the mode itself.
//
// It is a VALUE, and its ZERO VALUE IS THE SETTING MODE — TiltMachineNone is the zero of
// Wiring.TiltMachine. A Node built as a literal is therefore already in the mode a node starts in,
// with no constructor to call and no nil to test for. There is no second spelling of any mode, so
// two machines are equal exactly when they are in the same mode.
type tiltMachine struct {
	mode Wiring.TiltMachine
}

// homeSets is the per-mode data, and the ONLY per-mode thing there is: which separations that mode
// calls a resting state. A new mode is a new entry here — not a new type, not a new file, and not a
// branch anywhere in the rule below.
//
// A home is a position on the lattice, so each entry takes the ring: a quarter turn is 12 points on
// a 48-point ring and 6 on a 24-point one, and a node can change lattice underneath a running
// machine (Node.adoptLattice).
//
// perpendicular and parallel are not independent: quarter is the midpoint of 0 and half, which is
// why their two misses sum to a constant quarter and why each one's home is the other's farthest
// point (docs/pair-node/audit.html).
var homeSets = map[Wiring.TiltMachine]func(r *ring) []int32{
	// Setting: which machine to run is still being decided, so NOTHING IS OUT OF PLACE YET —
	// every separation is a resting state. That is what makes a node being set up hold still by
	// the ordinary rule instead of by an exemption from it: miss is zero wherever it stands, so
	// it is always halted, so step is never reached.
	Wiring.TiltMachineNone: func(r *ring) []int32 {
		// separation folds into [0, half], so this is every separation there is.
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

// homes is this machine's resting separations on the given ring — its mode's row in homeSets.
func (m tiltMachine) homes(r *ring) []int32 { return homeSets[m.mode](r) }

// miss is how far this arrival is from the NEAREST of this machine's resting states — zero when
// it is one of them. This is the whole per-mode difference, and it is a minimum over data: the
// function never learns which mode it is computing for.
func (m tiltMachine) miss(from, arrival *tiltState) int32 {
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
func (m tiltMachine) halted(from, arrival *tiltState) bool { return m.miss(from, arrival) == 0 }

// step is the single move — next or prev, a link either way, so it cannot leave the ring — that
// leaves the node closer to its halt. Where a mode has more than one resting state this closes on
// whichever is nearer, because miss already reports the nearest.
func (m tiltMachine) step(from, arrival *tiltState) *tiltState {
	if m.miss(from.next, arrival) <= m.miss(from.prev, arrival) {
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
// this package does not recognise has no home set, and running it would panic on the nil lookup
// rather than silently behaving like some other mode — so it is refused here, where the name of
// the broken invariant can still be said.
func machineFor(choice Wiring.TiltMachine) tiltMachine {
	if _, known := homeSets[choice]; !known {
		// The numeric value, not choice.String(): that falls back to "none" for anything it
		// does not recognise, so it would name the wrong mode in exactly this message.
		panic("Node1: no home set for tilt machine " + strconv.Itoa(int(choice)) +
			" — every mode must name the separations it rests at (machine.go homeSets)")
	}
	return tiltMachine{mode: choice}
}
