package runtopology

import (
	W "github.com/dtauraso/wirefold/src/Node/Wiring/dispatch"
	"github.com/dtauraso/wirefold/src/Node/Wiring/scene"
	"github.com/dtauraso/wirefold/src/Node/Wiring/scenecamera"
	"github.com/dtauraso/wirefold/src/Node/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/src/Node/Wiring/viewpersist"
	"github.com/dtauraso/wirefold/src/Chrome/SliderPanel"
	T "github.com/dtauraso/wirefold/src/Trace"
)

func loadSceneState(scenePath string, md *W.MoveDispatch, tr *T.Trace, speedSinks SliderPanel.Sinks) {
	scenecamera.SeedInitialViewpoint(scenePath, md.UI.VP.SetViewpoint, md.UI.VP.EmitViewpoint, tr)

	s := scene.For(scenePath)
	md.UI.SceneEditable = s.Editable
	md.UI.SceneKinds = s.KindMask()

	scenepersist.InstallOverlays(&md.UI, scenePath, tr)

	scenepersist.InstallPanels(&md.UI, scenePath)

	scenepersist.InstallSpeed(&md.UI, scenePath, speedSinks, tr)

	viewpersist.EnableViewpointPersist(&md.Persist, &md.UI, scenePath)

	viewpersist.EnableEditPersist(&md.Persist, &md.Scenes, &md.MR, scenePath)

	scenepersist.InstallSceneSphere(&md.UI, &md.GS, scenePath)
}
