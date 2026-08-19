package FitButton

import (
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/camera"
	"github.com/dtauraso/wirefold/nodes/spatial"
)

func FitDistance(fovDeg, aspect, width, height float64) float64 {
	fovRad := fovDeg * math.Pi / 180
	halfTan := math.Tan(fovRad / 2)
	return math.Max(height/2, width/2/aspect) / halfTan
}

func HomeFitPose(centers map[string]spatial.Vec3, radius map[string]float64, fovDeg, aspect float64) (pivot spatial.Vec3, r float64, pos, up camera.Dir, ok bool) {
	if len(centers) == 0 {
		return spatial.Vec3{}, 0, camera.Dir{}, camera.Dir{}, false
	}
	minX, minY, minZ := math.Inf(1), math.Inf(1), math.Inf(1)
	maxX, maxY, maxZ := math.Inf(-1), math.Inf(-1), math.Inf(-1)
	for id, p := range centers {
		rad := radius[id]
		minX, maxX = math.Min(minX, p.X-rad), math.Max(maxX, p.X+rad)
		minY, maxY = math.Min(minY, p.Y-rad), math.Max(maxY, p.Y+rad)
		minZ, maxZ = math.Min(minZ, p.Z-rad), math.Max(maxZ, p.Z+rad)
	}
	center := spatial.Vec3{X: (minX + maxX) / 2, Y: (minY + maxY) / 2, Z: (minZ + maxZ) / 2}
	sizeX, sizeY, sizeZ := maxX-minX, maxY-minY, maxZ-minZ
	dist := FitDistance(fovDeg, aspect, sizeX, sizeY) + sizeZ/2
	paddedDist := dist * 1.2
	pos = camera.WorldDirToAngles(spatial.Vec3{X: 0, Y: 0, Z: 1})
	up = camera.WorldDirToAngles(spatial.Vec3{X: 0, Y: 1, Z: 0})
	return center, paddedDist, pos, up, true
}
