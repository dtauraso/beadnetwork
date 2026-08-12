package geom

// camera_homefit.go — home fit (mirrors camera-ui.tsx HomeButton.onClick + geometry-helpers.ts
// boundingBox3D / fitDistance): frame all node bodies square-on. Split from camera_angles.go
// by concern (see that file's header for the shared quarantine rationale).

import "math"

// FitDistanceGo mirrors geometry-helpers.ts fitDistance: how far along the view axis to
// place the camera so a width×height world view fills the viewport at (fovDeg, aspect):
//
//	d = max(height/2, width/2/aspect) / tan(fov/2)
func FitDistanceGo(fovDeg, aspect, width, height float64) float64 {
	fovRad := fovDeg * math.Pi / 180
	halfTan := math.Tan(fovRad / 2)
	return math.Max(height/2, width/2/aspect) / halfTan
}

// HomeFitPose ports camera-ui.tsx HomeButton.onClick FORMULA-FAITHFULLY: build the AABB over
// node centers ± body radius (geometry-helpers.ts boundingBox3D), place the camera square-on
// in front of the content plane along +z with +y up, at a padded fit distance:
//
//	dist       = fitDistance(fov, aspect, sizeX, sizeY) + sizeZ/2
//	paddedDist = dist * 1.2
//	pivot = bbox center; r = paddedDist; pos = dir(+z); up = dir(+y)
//
// centers maps node id → world center; radius maps node id → body sphere radius. Returns
// ok=false when there are no nodes (HomeButton returns early in that case). pos/up come from
// WorldDirToAngles so the resulting eye = pivot + r·(+z) and screen-up = +y, matching the
// TS cam.position / cam.up / cam.lookAt(center).
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
