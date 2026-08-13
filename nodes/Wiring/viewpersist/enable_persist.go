package viewpersist

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/moverreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/sceneswitch"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
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
	// An edge owns its own file, so it is armed the same way its endpoints are.
	for _, em := range mr.EdgeMovers() {
		em.SetPersistRoot(root)
	}
}
