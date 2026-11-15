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

	SpeedSlider bool

	AnglesPill bool

	PlusButton bool

	OverlaysDock bool

	HiddenOverlays []string

	RemovedOverlays []string

	FixedOverlays []string
}

var All = []Scene{

	{Name: "ring", Dir: "topology", CoplanarEdges: false, UpAxis: false, ClockDivisor: 1, Editable: true, NodesDraggable: true, SpeedSlider: true, AnglesPill: true, PlusButton: true, OverlaysDock: true,
		Kinds: []string{
			"Input", "Time", "TimeStart", "TimeEnd",
			"Pulse", "PulseLeft", "PulseRight",
			"SelectLeft", "SelectRight",
		},
		RemovedOverlays: []string{"nodePoles", "scenePoles", "allPoleSpheres"},
		FixedOverlays:   []string{"nodeBody", "nodeRing", "edgeVectors"}},

	{Name: "φ, θ", Dir: "topology-phi-theta", CoplanarEdges: true, UpAxis: false, ClockDivisor: 8, Editable: true, NodesDraggable: true, Kinds: []string{"PhiTheta"}, PlusButton: true,
		OverlaysDock: true, HiddenOverlays: []string{"edgeVectors"},
		RemovedOverlays: []string{"nodePoles", "scenePoles", "allPoleSpheres"},
		FixedOverlays:   []string{"nodeBody", "nodeRing"}},
}

var Unlisted = Scene{ClockDivisor: 1, SpeedSlider: true, AnglesPill: true}

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
