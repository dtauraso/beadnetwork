package edgegeom

func EdgeSegment(srcCenter, tgtCenter Vec3, srcOuterR, tgtOuterR float64) segment {
	dir := tgtCenter.Sub(srcCenter)
	if dir.Length() < 1e-9 {

		return segment{Start: srcCenter, End: tgtCenter}
	}
	unit := dir.Normalize()
	start := srcCenter.Add(unit.Scale(srcOuterR))
	end := tgtCenter.Sub(unit.Scale(tgtOuterR))
	return segment{Start: start, End: end}
}

func EdgeCenterDistAndDir(selfCenter, targetCenter Vec3) (dist float64, unitDir Vec3, ok bool) {
	delta := targetCenter.Sub(selfCenter)
	length := delta.Length()
	if length < 1e-9 {
		return 0, Vec3{}, false
	}
	return length, delta.Normalize(), true
}
