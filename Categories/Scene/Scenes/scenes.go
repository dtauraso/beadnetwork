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

	NodesDraggable bool

	Kinds []string

	PolarRules bool

	SpeedSlider bool

	AnglesPill bool

	PlusButton bool
}

var All = []Scene{

	{Name: "ring", Dir: "topology", CoplanarEdges: false, UpAxis: false, ClockDivisor: 1, Editable: true, NodesDraggable: true, PolarRules: true, SpeedSlider: true, AnglesPill: true,
		Kinds: []string{
			"Input", "Time", "TimeStart", "TimeEnd",
			"Pulse", "PulseLeft", "PulseRight",
			"SelectLeft", "SelectRight",
		}},

	{Name: "pair φ 2", Dir: "topology-pair", CoplanarEdges: true, UpAxis: true, ClockDivisor: 64, Editable: true, NodesDraggable: true, Kinds: []string{"NodePhi"}, PolarRules: true, SpeedSlider: true, AnglesPill: true},

	{Name: "pair φ, θ 2", Dir: "topology-pair-phi-theta", CoplanarEdges: true, UpAxis: false, ClockDivisor: 8, Editable: true, NodesDraggable: false, Kinds: []string{"NodePhiTheta"}, PolarRules: true, SpeedSlider: true, AnglesPill: true},

	{Name: "pair φ, θ 3", Dir: "topology-pair-phi-theta-3", CoplanarEdges: true, UpAxis: false, ClockDivisor: 8, Editable: true, NodesDraggable: false, Kinds: []string{"NodePhiTheta3"}, PlusButton: true},
}

var Unlisted = Scene{ClockDivisor: 1, PolarRules: true, SpeedSlider: true, AnglesPill: true}

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
