package Scene

type SceneValue struct {
	Name string
	Path string
	Kind string
}

var SceneValues = []SceneValue{
	{Name: "cx", Path: "view/sphere/cx.bin", Kind: "f64"},
	{Name: "cy", Path: "view/sphere/cy.bin", Kind: "f64"},
	{Name: "cz", Path: "view/sphere/cz.bin", Kind: "f64"},
	{Name: "radius", Path: "view/sphere/radius.bin", Kind: "f64"},
	{Name: "constantR", Path: "constants/constant-r.bin", Kind: "f64"},
	{Name: "maxIndexPhi", Path: "constants/max-index-phi.bin", Kind: "i64"},
	{Name: "maxIndexTheta", Path: "constants/max-index-theta.bin", Kind: "i64"},

	{Name: "spawn", Path: "view/spawn.bin", Kind: "i64"},
}

func SceneValuePath(sceneRoot, name string) string {
	for _, v := range SceneValues {
		if v.Name == name {
			return sceneRoot + "/" + v.Path
		}
	}
	panic("Scene.SceneValuePath: " + name + " is not a scene value, so Go and the renderer disagree about where it lives; add it to SceneValues in src/Scene/scene_values.go")
}
