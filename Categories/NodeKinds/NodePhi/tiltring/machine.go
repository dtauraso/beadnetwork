package tiltring

import (
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
)

type Machine struct {
	Mode TiltPanel.TiltMachine
}

func (m Machine) OffsetOn(r *Ring, top, arrival int32) int32 {
	if m.Mode == TiltPanel.TiltMachineNone {
		return 0
	}
	return Offset(top, arrival, r.Points)
}

func (m Machine) Choice() TiltPanel.TiltMachine { return m.Mode }

func (m Machine) String() string {
	if m.Mode == TiltPanel.TiltMachineNone {
		return "unset"
	}
	return m.Mode.String()
}

func Unset() Machine { return Machine{Mode: TiltPanel.TiltMachineNone} }

var knownMachines = map[TiltPanel.TiltMachine]bool{
	TiltPanel.TiltMachineNone:     true,
	TiltPanel.TiltMachineParallel: true,
}

func MachineFor(choice TiltPanel.TiltMachine) Machine {
	if !knownMachines[choice] {
		panic("tiltring: no rule for tilt machine " + strconv.Itoa(int(choice)) +
			" — a machine must be one this package knows how to turn (machine.go knownMachines)")
	}
	return Machine{Mode: choice}
}
