package scene

type SceneTab struct {
	Name string
	Dir  string

	QuantizedDrag bool

	CoplanarEdges bool

	UpAxis bool

	ClockDivisor float64

	DistanceGroups bool

	Editable bool

	Kinds []string
}

var SceneTabs = []SceneTab{

	{Name: "ring", Dir: "topology", QuantizedDrag: false, CoplanarEdges: false, UpAxis: false, ClockDivisor: 1, DistanceGroups: true, Editable: true,
		Kinds: []string{
			"Input", "Time", "TimeStart", "TimeEnd",
			"Pulse", "PulseLeft", "PulseRight",
			"SelectLeft", "SelectRight",
			"HoldFlip", "Pacer",
		}},

	{Name: "pair", Dir: "topology-pair", QuantizedDrag: false, CoplanarEdges: true, UpAxis: true, ClockDivisor: 64, DistanceGroups: false, Editable: true, Kinds: []string{"PairNode", "NormalSum"}},
}
