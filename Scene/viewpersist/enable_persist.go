package viewpersist

import (
	"github.com/dtauraso/wirefold/Node/moverreg"
	"github.com/dtauraso/wirefold/Scene/sceneswitch"
	"github.com/dtauraso/wirefold/Scene/viewstate"
)

func EnableViewpointPersist(persist *Persisters, ui *viewstate.UIState, topologyPath string) {
	p := persist.ArmViewpoint(topologyPath)
	ui.VP.Persist = p.Schedule
}

func EnableEditPersist(persist *Persisters, scenes *sceneswitch.SceneSwitch, mr *moverreg.MoverRegistry, topologyPath string) {
	root := topologyPath

	scenes.TreeRoot = root
	persist.ArmEdit(topologyPath)

	for _, nm := range mr.NodeGeoms() {
		nm.SetPersistRoot(root)
	}
}
