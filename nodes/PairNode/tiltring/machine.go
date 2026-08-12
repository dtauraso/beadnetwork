package tiltring

import (
	"strconv"

	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

type Machine struct {
	Mode tiltvector.TiltMachine
}

func (s *State) NearerEndCount(arrival *State) (c int32, atBottom bool) {
	u := ((s.Idx-arrival.Idx)%s.R.Points + s.R.Points) % s.R.Points
	if u < s.R.HalfTurn {
		return u, false
	}

	return ((s.Opposite.Idx-arrival.Idx)%s.R.Points + s.R.Points) % s.R.Points, true
}

type StoppingCount struct {
	Anywhere bool

	At func(r *Ring) int32
}

var stoppingCounts = map[tiltvector.TiltMachine]StoppingCount{

	tiltvector.TiltMachineNone: {Anywhere: true},

	tiltvector.TiltMachinePerpendicular: {At: func(r *Ring) int32 { return 0 }},

	tiltvector.TiltMachineParallel: {At: func(r *Ring) int32 { return r.QuarterTurn }},
}

func (m Machine) Stopping() StoppingCount { return stoppingCounts[m.Mode] }

func (m Machine) Settled(from, arrival *State) bool {
	stop := m.Stopping()
	if stop.Anywhere {
		return true
	}
	c, _ := from.NearerEndCount(arrival)
	return c == stop.At(from.R)
}

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

func (m Machine) Choice() tiltvector.TiltMachine { return m.Mode }

func (m Machine) String() string {
	if m.Mode == tiltvector.TiltMachineNone {
		return "setting"
	}
	return m.Mode.String()
}

var Setting = Machine{Mode: tiltvector.TiltMachineNone}

func MachineFor(choice tiltvector.TiltMachine) Machine {
	if _, known := stoppingCounts[choice]; !known {

		panic("tiltring: no stopping counts for tilt machine " + strconv.Itoa(int(choice)) +
			" — every mode must name the counts it stops at (machine.go stoppingCounts)")
	}
	return Machine{Mode: choice}
}
