package owners

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
