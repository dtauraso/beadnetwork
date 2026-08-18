package runtopology

import (
	W "github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenecamera"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewpersist"
	"github.com/dtauraso/wirefold/tools/topology-vscode/SliderPanel"
	T "github.com/dtauraso/wirefold/tools/topology-vscode/Trace"
)

func loadSceneState(scenePath string, md *W.MoveDispatch, tr *T.Trace, speedSinks SliderPanel.Sinks) {
	scenecamera.SeedInitialViewpoint(scenePath, md.UI.VP.SetViewpoint, md.UI.VP.EmitViewpoint, tr)

	scenepersist.InstallOverlays(&md.UI, scenePath, tr)

	scenepersist.InstallPanels(&md.UI, scenePath)

	scenepersist.InstallSpeed(&md.UI, scenePath, speedSinks, tr)

	viewpersist.EnableViewpointPersist(&md.Persist, &md.UI, scenePath)

	viewpersist.EnableEditPersist(&md.Persist, &md.Scenes, &md.MR, scenePath)

	scenepersist.InstallSceneSphere(&md.UI, &md.GS, scenePath)
}
