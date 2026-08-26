package Scenes

import (
	"path/filepath"
)

type Scene struct {
	Name string

	Dir string

	CoplanarEdges bool

	UpAxis bool

	ClockDivisor float64

	Editable bool

	Kinds []string
}

var All = []Scene{

	{Name: "ring", Dir: "topology", CoplanarEdges: false, UpAxis: false, ClockDivisor: 1, Editable: true,
		Kinds: []string{
			"Input", "Time", "TimeStart", "TimeEnd",
			"Pulse", "PulseLeft", "PulseRight",
			"SelectLeft", "SelectRight",
		}},

	{Name: "pair φ", Dir: "topology-pair", CoplanarEdges: true, UpAxis: true, ClockDivisor: 64, Editable: true, Kinds: []string{"NodePhi"}},

	{Name: "pair φ, θ", Dir: "topology-pair-phi-theta", CoplanarEdges: true, UpAxis: true, ClockDivisor: 64, Editable: true, Kinds: []string{"NodePhiTheta"}},
}

var Unlisted = Scene{ClockDivisor: 1}

func Declared(path string) (Scene, bool) {
	base := filepath.Base(filepath.Clean(path))
	for _, s := range All {
		if s.Dir == base {
			return s, true
		}
	}
	return Scene{}, false
}

func For(path string) Scene {
	if s, ok := Declared(path); ok {
		return s
	}
	return Unlisted
}
