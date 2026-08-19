package scene

import (
	NodeBuf "github.com/dtauraso/wirefold/src/Node"
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
			"HoldFlip", "Pacer",
		}},

	{Name: "pair", Dir: "topology-pair", CoplanarEdges: true, UpAxis: true, ClockDivisor: 64, Editable: true, Kinds: []string{"PairNode", "NormalSum"}},
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

func (s Scene) KindMask() uint32 {
	if len(s.Kinds) == 0 {
		return ^uint32(0)
	}
	var mask uint32
	for _, k := range s.Kinds {
		if id := NodeBuf.NodeKindID(k); id != NodeBuf.KindIDUnknown {
			mask |= 1 << uint(id)
		}
	}
	return mask
}
