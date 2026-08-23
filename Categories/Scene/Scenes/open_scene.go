package Scenes

type SceneSwitch struct {
	AnchorPath string
	Quit       func()

	TreeRoot string

	Loaded int
}
