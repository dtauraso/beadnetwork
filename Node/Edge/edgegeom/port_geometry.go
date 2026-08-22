package edgegeom

import (
	"github.com/dtauraso/wirefold/Node/nodegeom"
)

func EdgeSegment(src, tgt nodegeom.NodeGeom) segment {
	srcCenter := nodegeom.NodeWorldPos(src)
	tgtCenter := nodegeom.NodeWorldPos(tgt)
	dir := tgtCenter.Sub(srcCenter)
	if dir.Length() < 1e-9 {

		return segment{Start: Vec3(srcCenter), End: Vec3(tgtCenter)}
	}
	unit := dir.Normalize()
	start := srcCenter.Add(unit.Scale(nodegeom.NodeTorusOuterR(src.Kind)))
	end := tgtCenter.Sub(unit.Scale(nodegeom.NodeTorusOuterR(tgt.Kind)))
	return segment{Start: Vec3(start), End: Vec3(end)}
}

func EdgeCenterDistAndDir(selfCenter, targetCenter Vec3) (dist float64, unitDir Vec3, ok bool) {
	delta := targetCenter.Sub(selfCenter)
	length := delta.Length()
	if length < 1e-9 {
		return 0, Vec3{}, false
	}
	return length, delta.Normalize(), true
}
