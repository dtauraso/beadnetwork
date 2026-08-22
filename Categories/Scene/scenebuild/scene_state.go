package scenebuild

import (
	"github.com/dtauraso/wirefold/Categories/Camera"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Scene/scene"
	"github.com/dtauraso/wirefold/Categories/Scene/scenepersist"
	"github.com/dtauraso/wirefold/Categories/Scene/scenerun"
	"github.com/dtauraso/wirefold/Categories/Scene/viewpersist"
)

func LoadSceneState(scenePath string, md *scenerun.MoveDispatch, speedSinks SliderPanel.Sinks) {
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
