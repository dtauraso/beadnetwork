package nodeactor

func (f *sceneFlags) SetSceneFlags(coplanarEdges, upAxis bool) {
	f.coplanarEdges = coplanarEdges
	f.upAxis = upAxis
}

func (f *sceneFlags) Flags() (coplanarEdges, upAxis bool) {
	return f.coplanarEdges, f.upAxis
}
