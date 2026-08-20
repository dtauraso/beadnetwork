package runtopology

import (
	"github.com/dtauraso/wirefold/src/Chrome/SliderPanel"
	W "github.com/dtauraso/wirefold/src/Node/Wiring/dispatch"
	"github.com/dtauraso/wirefold/src/Node/Wiring/scene"
	"github.com/dtauraso/wirefold/src/Node/Wiring/scenecamera"
	"github.com/dtauraso/wirefold/src/Node/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/src/Node/Wiring/viewpersist"
)

func loadSceneState(scenePath string, md *W.MoveDispatch, speedSinks SliderPanel.Sinks) {
	scenecamera.SeedInitialViewpoint(scenePath, md.UI.VP.SetViewpoint, md.UI.VP.EmitViewpoint)

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
