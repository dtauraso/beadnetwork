package sceneswitch

type SceneSwitch struct {
	AnchorPath string
	Quit       func()

	TreeRoot string
}
