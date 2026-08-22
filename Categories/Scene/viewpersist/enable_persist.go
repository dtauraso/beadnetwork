package viewpersist

import (
	"github.com/dtauraso/wirefold/Categories/Node/moverreg"
	"github.com/dtauraso/wirefold/Categories/Scene/sceneswitch"
	"github.com/dtauraso/wirefold/Categories/Scene/viewstate"
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
