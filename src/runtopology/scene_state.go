package runtopology

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	W "github.com/dtauraso/wirefold/src/Input/dispatch"
	"github.com/dtauraso/wirefold/src/Scene/scene"
	"github.com/dtauraso/wirefold/src/Camera"
	"github.com/dtauraso/wirefold/src/Scene/scenepersist"
	"github.com/dtauraso/wirefold/src/Node/Wiring/viewpersist"
)

func loadSceneState(scenePath string, md *W.MoveDispatch, speedSinks SliderPanel.Sinks) {
	Camera.SeedInitialViewpoint(scenePath, md.UI.VP.SetViewpoint, md.UI.VP.EmitViewpoint)

	s := scene.For(scenePath)
	md.UI.SceneEditable = s.Editable
	md.UI.SceneKinds = s.KindMask()

	scenepersist.InstallOverlays(&md.UI, scenePath)

	scenepersist.InstallPanels(&md.UI, scenePath)

	scenepersist.InstallSpeed(&md.UI, scenePath, speedSinks)

	viewpersist.EnableViewpointPersist(&md.Persist, &md.UI, scenePath)

	viewpersist.EnableEditPersist(&md.Persist, &md.Scenes, &md.MR, scenePath)

	scenepersist.InstallSceneSphere(&md.UI, &md.GS, scenePath)
}
