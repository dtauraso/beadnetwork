package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Polar/polarindex"
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
	err := WriteDragIndex(nm.persistRoot, nm.id, off.Phi, off.Theta, off.R, nm.tilt.TopTiltVectorPhiIdx())
	if err != nil {
		LogPersistErr("index_persist", nm.id, err)
	}
}
