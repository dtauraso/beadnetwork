package tiltring

import (
	"strconv"

	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

type Machine struct {
	Mode tiltvector.TiltMachine
}

func (m Machine) OffsetOn(r *Ring, top, arrival int32) int32 {
	if m.Mode == tiltvector.TiltMachineNone {
		return 0
	}
	return Offset(top, arrival, r.Points)
}

func (m Machine) Choice() tiltvector.TiltMachine { return m.Mode }

func (m Machine) String() string {
	if m.Mode == tiltvector.TiltMachineNone {
		return "unset"
	}
	return m.Mode.String()
}

func Unset() Machine { return Machine{Mode: tiltvector.TiltMachineNone} }

var knownMachines = map[tiltvector.TiltMachine]bool{
	tiltvector.TiltMachineNone:     true,
	tiltvector.TiltMachineParallel: true,
}

func MachineFor(choice tiltvector.TiltMachine) Machine {
	if !knownMachines[choice] {
		panic("tiltring: no rule for tilt machine " + strconv.Itoa(int(choice)) +
			" — a machine must be one this package knows how to turn (machine.go knownMachines)")
	}
	return Machine{Mode: choice}
}
