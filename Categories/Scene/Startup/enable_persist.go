package Startup

import (
	"github.com/dtauraso/wirefold/Categories/Node"
	"github.com/dtauraso/wirefold/Categories/Scene/Scenes"
	"github.com/dtauraso/wirefold/Categories/Scene/View"
)

func EnableViewpointPersist(ui *View.UIState, topologyPath string) {
	ui.VP.Persist = armViewpoint(topologyPath).Schedule
}

func EnableEditPersist(ui *View.UIState, scenes *Scenes.SceneSwitch, nodeGeoms map[string]*Node.NodeGeometry, topologyPath string) {
	scenes.TreeRoot = topologyPath
	armEdit(ui, topologyPath)

	for _, nm := range nodeGeoms {
		nm.SetPersistRoot(topologyPath)
	}
}
