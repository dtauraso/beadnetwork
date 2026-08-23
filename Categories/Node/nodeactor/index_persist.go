package nodeactor

import (
	"github.com/dtauraso/wirefold/Categories/Node/nodeactor/nodefiles"
	"github.com/dtauraso/wirefold/Categories/Polar/polarindex"
)

func (nm *NodeGeometry) persistIndex(off polarindex.Offset) {
	nm.writeIndex(off)
}

func (nm *NodeGeometry) persistTiltVectorAngle() {
	nm.writeIndex(nm.geom.DragIndex)
}

func (nm *NodeGeometry) writeIndex(off polarindex.Offset) {
	if nm.persistRoot == "" {
		return
	}
	err := nodefiles.WriteDragIndex(nm.persistRoot, nm.id, off.Phi, off.Theta, off.R, nm.tilt.TopTiltVectorPhiIdx())
	if err != nil {
		nodefiles.LogPersistErr("index_persist", nm.id, err)
	}
}
