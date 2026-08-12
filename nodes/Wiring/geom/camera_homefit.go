package geom

import "math"

func FitDistanceGo(fovDeg, aspect, width, height float64) float64 {
	fovRad := fovDeg * math.Pi / 180
	halfTan := math.Tan(fovRad / 2)
	return math.Max(height/2, width/2/aspect) / halfTan
}

func HomeFitPose(centers map[string]vec3, radius map[string]float64, fovDeg, aspect float64) (pivot vec3, r float64, pos, up Dir, ok bool) {
	if len(centers) == 0 {
		return vec3{}, 0, Dir{}, Dir{}, false
	}
	minX, minY, minZ := math.Inf(1), math.Inf(1), math.Inf(1)
	maxX, maxY, maxZ := math.Inf(-1), math.Inf(-1), math.Inf(-1)
	for id, p := range centers {
		rad := radius[id]
		minX, maxX = math.Min(minX, p.X-rad), math.Max(maxX, p.X+rad)
		minY, maxY = math.Min(minY, p.Y-rad), math.Max(maxY, p.Y+rad)
		minZ, maxZ = math.Min(minZ, p.Z-rad), math.Max(maxZ, p.Z+rad)
	}
	center := vec3{X: (minX + maxX) / 2, Y: (minY + maxY) / 2, Z: (minZ + maxZ) / 2}
	sizeX, sizeY, sizeZ := maxX-minX, maxY-minY, maxZ-minZ
	dist := FitDistanceGo(fovDeg, aspect, sizeX, sizeY) + sizeZ/2
	paddedDist := dist * 1.2
	pos = WorldDirToAngles(vec3{X: 0, Y: 0, Z: 1})
	up = WorldDirToAngles(vec3{X: 0, Y: 1, Z: 0})
	return center, paddedDist, pos, up, true
}
