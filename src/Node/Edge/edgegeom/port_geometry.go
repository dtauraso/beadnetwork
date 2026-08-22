package edgegeom

import (
	"github.com/dtauraso/wirefold/src/Node/nodegeom"
	"github.com/dtauraso/wirefold/src/spatial"
)

func EdgeSegment(src, tgt nodegeom.NodeGeom) segment {
	srcCenter := nodegeom.NodeWorldPos(src)
	tgtCenter := nodegeom.NodeWorldPos(tgt)
	dir := tgtCenter.Sub(srcCenter)
	if dir.Length() < 1e-9 {

		return segment{Start: srcCenter, End: tgtCenter}
	}
	unit := dir.Normalize()
	start := srcCenter.Add(unit.Scale(nodegeom.NodeTorusOuterR(src.Kind)))
	end := tgtCenter.Sub(unit.Scale(nodegeom.NodeTorusOuterR(tgt.Kind)))
	return segment{Start: start, End: end}
}

func EdgeCenterDistAndDir(selfCenter, targetCenter spatial.Vec3) (dist float64, unitDir spatial.Vec3, ok bool) {
	delta := targetCenter.Sub(selfCenter)
	length := delta.Length()
	if length < 1e-9 {
		return 0, spatial.Vec3{}, false
	}
	return length, delta.Normalize(), true
}
