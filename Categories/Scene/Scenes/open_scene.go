package Scenes

type SceneSwitch struct {
	AnchorPath string

	Open func(idx int)

	Reload func()

	TreeRoot string

	NodesRoot string

	Loaded int
}
