package Scenes

type SceneSwitch struct {
	AnchorPath string
	Quit       func()

	Open func(idx int)

	TreeRoot string

	Loaded int
}
