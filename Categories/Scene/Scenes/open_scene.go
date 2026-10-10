package Scenes

type SceneSwitch struct {
	AnchorPath string

	Open func(idx int)

	Reload func()

	TreeRoot string

	Loaded int
}
