// Package sceneswitch holds the tab-switch owner split out of MoveDispatch (god-object
// decomposition), as a pure move (no logic changes): SceneSwitch is MoveDispatch's half
// of scene-tab switching — the anchor to persist against, and the way to end this process
// so the runner's looping respawn loads the newly selected scene. Both AnchorPath/Quit are
// zero/nil until runtopology/topology_run.go sets md.Scenes.AnchorPath/md.Scenes.Quit
// directly (the exported field, no method — EnableSceneSwitch, a pure two-field forward,
// was deleted), so a bare test-constructed MoveDispatch cannot exit anything.
//
// The surrounding logic (SceneTabs, SelectScene, structural-edit handling) stays in
// package Wiring: it is genuine orchestration referencing Wiring-only state (SceneTabs,
// persistence helpers), not a thin delegator to this type, so only the plain data this
// struct holds moves here.
package sceneswitch

// SceneSwitch is MoveDispatch's half of scene-tab switching: the anchor to persist the
// selection against, the quit func whose call the extension host's looping respawn
// follows, and the loaded scene's own tree root for structural edits.
type SceneSwitch struct {
	AnchorPath string
	Quit       func()
	// TreeRoot is the LOADED scene's own directory (the anchor's sibling the selected tab
	// points at), set by EnableEditPersist. A structural edit writes here, while the tab
	// SELECTION is written at the anchor — the two are different paths for the reason
	// scene_tabs.go's header gives: a selection stored inside the scene it selects is
	// unreachable while another scene is loaded.
	TreeRoot string
}
