package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
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
	sc := nm.Constants()
	err := WriteDragIndex(nm.persistRoot, nm.id, off.Phi, off.Theta, off.R, nm.tilt.TopTiltVectorPhiIdx(), sc.MaxIndexPhi, sc.MaxIndexTheta)
	if err != nil {
		LogPersistErr("index_persist", nm.id, err)
	}
}
