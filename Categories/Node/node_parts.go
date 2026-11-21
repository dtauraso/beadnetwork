package Node

type NodeData struct {
	Label  string
	Init   []int
	Repeat bool
	State  map[string]int

	SendRules map[string]string
}

type Flags struct {
	coplanarEdges bool
	upAxis        bool
}

func (f *Flags) SetSceneFlags(coplanarEdges, upAxis bool) {
	f.coplanarEdges = coplanarEdges
	f.upAxis = upAxis
}

func (f *Flags) Flags() (coplanarEdges, upAxis bool) {
	return f.coplanarEdges, f.upAxis
}

type UI struct {
	selected, hovered, latchedSel uint8
	hoverPort                     string
	hoverIsInput                  bool
}

func (u *UI) SetSelected(on bool) {
	if on {
		u.selected = 1
	} else {
		u.selected = 0
	}
}

func (u *UI) SetHover(on bool, port string, isInput bool) {
	if on {
		u.hovered = 1
		u.hoverPort = port
		u.hoverIsInput = isInput
	} else {
		u.hovered = 0
		u.hoverPort = ""
		u.hoverIsInput = false
	}
}

func (u *UI) Flags() (selected, latchedSel uint8) {
	return u.selected, u.latchedSel
}

func (u *UI) SetLatched(on bool) {
	if on {
		u.latchedSel = 1
	} else {
		u.latchedSel = 0
	}
}

type Topology struct {
	edgeIDs       []string
	neighborKinds map[string]string
	mutualTargets map[string]bool
	nodeRowFor    func(id string) (int32, bool)
}

func NewTopology() Topology {
	return Topology{}
}

func (t *Topology) EdgeIDs() []string { return t.edgeIDs }

func (t *Topology) NeighborKinds() map[string]string { return t.neighborKinds }

func (t *Topology) AddMutualTarget(target string) {
	if t.mutualTargets == nil {
		t.mutualTargets = map[string]bool{}
	}
	t.mutualTargets[target] = true
}

func (t *Topology) AddEdgeID(edgeID string) {
	t.edgeIDs = append(t.edgeIDs, edgeID)
}

func (t *Topology) AddNeighborKind(toID, kind string) {
	if t.neighborKinds == nil {
		t.neighborKinds = map[string]string{}
	}
	t.neighborKinds[toID] = kind
}

func (t *Topology) NeighborKind(toID string) string {
	return t.neighborKinds[toID]
}

func (t *Topology) IsMutualTarget(toID string) bool {
	return t.mutualTargets[toID]
}

func (t *Topology) SetNodeRowFor(fn func(id string) (int32, bool)) {
	t.nodeRowFor = fn
}

func (t *Topology) NodeRowFor(id string) (int32, bool) {
	if t.nodeRowFor == nil {
		return -1, false
	}
	return t.nodeRowFor(id)
}
