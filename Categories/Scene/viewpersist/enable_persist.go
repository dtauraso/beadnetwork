package viewpersist

import (
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor"
	"github.com/dtauraso/wirefold/Categories/Scene/sceneswitch"
	"github.com/dtauraso/wirefold/Categories/Scene/viewstate"
)

func EnableViewpointPersist(persist *Persisters, ui *viewstate.UIState, topologyPath string) {
	p := persist.ArmViewpoint(topologyPath)
	ui.VP.Persist = p.Schedule
}

func EnableEditPersist(persist *Persisters, scenes *sceneswitch.SceneSwitch, nodeGeoms map[string]*nodeactor.NodeGeometry, topologyPath string) {
	root := topologyPath

	scenes.TreeRoot = root
	persist.ArmEdit(topologyPath)

	for _, nm := range nodeGeoms {
		nm.SetPersistRoot(root)
	}
}
