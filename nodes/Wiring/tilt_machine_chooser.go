package Wiring

// tilt_machine_chooser.go — WHICH STATE MACHINE THE PAIR RUNS IS CHOSEN OUTSIDE THE NODES,
// once per tilt click, from the gap between the two tilts.
//
// WHY IT CANNOT BE CHOSEN INSIDE A NODE. At the moment of a ▲/▼ click nothing has arrived on
// the vector channel yet — the exchange has not been started — so a node knows its own tilt and
// nothing about its partner's. Every attempt to pick the machine from inside ended up inferring
// it from the FIRST ARRIVAL instead, which is a different question asked at a different time,
// and always answered perpendicular: a node with nothing to return to closes on the arrival,
// and closing on the arrival IS the perpendicular measure. A pair set up close to perpendicular
// could never be asked to run the parallel machine.
//
// WHAT SEES BOTH TILTS. The view-owner goroutine routes every tilt edit (applyUpdateTiltVector,
// stdin_reader.go), so between a reset and a start it is the SOLE source of every increment
// either node has taken, and its own tally of them is exact. It is not a second copy of an
// authority: the node still owns its index and moves it during the exchange; this counts what
// the USER asked for while setting up, which is the only window the choice is made in.
//
// THE CHOICE STICKS. Once picked, the pair runs that machine until RESET, which drops the
// choice back to none along with the tallies.

// TiltMachine names which state machine a pair node runs. The pair kind maps these to its own
// two machines (nodes/Node1/perpendicular.go, parallel.go); this package names them without
// knowing anything about how either one steps.
type TiltMachine int8

const (
	// TiltMachineNone: run neither — what a reset restores.
	TiltMachineNone TiltMachine = iota
	// TiltMachinePerpendicular: the two tilts a quarter turn apart.
	TiltMachinePerpendicular
	// TiltMachineParallel: the two tilts pointing the same way.
	TiltMachineParallel
)

func (m TiltMachine) String() string {
	switch m {
	case TiltMachinePerpendicular:
		return "perpendicular"
	case TiltMachineParallel:
		return "parallel"
	}
	return "none"
}

// tiltMachineChooser is the view owner's own record of where the user has put each pair tilt,
// and the choice it made from them. Owned by that one goroutine — seeded on its build pass,
// read and written only from applyUpdateTiltVector — so there is nothing here to coordinate.
type tiltMachineChooser struct {
	// idx is each pair node's tilt index as the USER has set it: the persisted seed the node
	// built with, plus one per click routed to it since.
	idx map[string]int32
}

// seed records a pair node's load-time tilt index, called as that node claims its tilt-edit
// channel (BuildArgs.TiltEditIn) so the tally starts where the node itself starts.
func (c *tiltMachineChooser) seed(id string, thetaIdx int32) {
	if c.idx == nil {
		c.idx = map[string]int32{}
	}
	c.idx[id] = thetaIdx
}

// click applies one ▲/▼ to the tally — the same single step the node itself is taking.
func (c *tiltMachineChooser) click(id string, up bool, points int32) {
	if c.idx == nil || points <= 0 {
		return
	}
	next := c.idx[id] + 1
	if !up {
		next = c.idx[id] - 1
	}
	if next >= points {
		next -= points
	}
	if next < 0 {
		next += points
	}
	c.idx[id] = next
}

// forget returns every tally to zero — RESET, the one thing that releases the choice. A reset
// zeroes both tilts (Node1's clear), so the tally follows them to zero.
//
// IT DOES NOT DROP THE ENTRIES. Which node ids are in here is the list of pair nodes, written
// once at build time and never again; deleting them leaves the chooser with fewer than the two
// tilts it needs to compare, so it answers "none" to every click after the first reset and no
// machine is ever chosen. That froze a pair: eleven clicks and a start, and not one row in the
// log, because with no machine an arrival moves nothing.
func (c *tiltMachineChooser) forget() {
	for id := range c.idx {
		c.idx[id] = 0
	}
}

// choose classifies the CURRENT gap between the pair's two tilts:
//
//	the gap is a quarter turn  ->  the pair is perpendicular  ->  the perpendicular machine
//	anything else              ->  the gap is acute           ->  the parallel machine
//
// The gap is folded to the short way round, so it never exceeds a half turn and means the same
// thing whichever node is subtracted from which.
//
// It answers TiltMachineNone when there are not exactly two tilts to compare — a scene with one
// pair node or none is not a pair, and guessing for it would be worse than saying nothing.
func (c *tiltMachineChooser) choose(points int32) TiltMachine {
	if len(c.idx) != 2 || points <= 0 {
		return TiltMachineNone
	}
	var a, b int32
	first := true
	for _, v := range c.idx {
		if first {
			a, first = v, false
			continue
		}
		b = v
	}
	gap := a - b
	if gap < 0 {
		gap = -gap
	}
	if gap > points/2 {
		gap = points - gap
	}
	if gap == points/4 {
		return TiltMachinePerpendicular
	}
	return TiltMachineParallel
}
