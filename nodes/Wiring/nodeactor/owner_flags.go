package nodeactor

func (f *sceneFlags) SetSceneFlags(coplanarEdges, upAxis bool) {
	f.coplanarEdges = coplanarEdges
	f.upAxis = upAxis
}
