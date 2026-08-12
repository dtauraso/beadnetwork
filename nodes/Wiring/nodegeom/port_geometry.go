package nodegeom

func EdgeSegment(src, tgt NodeGeom) wireSegment {
	srcCenter := NodeWorldPos(src)
	tgtCenter := NodeWorldPos(tgt)
	dir := tgtCenter.Sub(srcCenter)
	if dir.Length() < 1e-9 {

		return wireSegment{Start: srcCenter, End: tgtCenter}
	}
	unit := dir.Normalize()
	start := srcCenter.Add(unit.Scale(NodeTorusOuterR(src.Kind)))
	end := tgtCenter.Sub(unit.Scale(NodeTorusOuterR(tgt.Kind)))
	return wireSegment{Start: start, End: end}
}

func EdgeCenterDistAndDir(selfCenter, targetCenter vec3) (dist float64, unitDir vec3, ok bool) {
	delta := targetCenter.Sub(selfCenter)
	length := delta.Length()
	if length < 1e-9 {
		return 0, vec3{}, false
	}
	return length, delta.Normalize(), true
}
