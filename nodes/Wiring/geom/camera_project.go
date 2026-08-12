package geom

// camera_project.go — projection (mirrors THREE.Vector3.project for the perspective
// camera) — used ONLY by zoom-to-cursor to find the node nearest the pointer in NDC. Small
// NDC error only changes which node is picked (the dolly is floored so it never reaches the
// target). Split from camera_angles.go by concern (see that file's header for the shared
// quarantine rationale).

import "math"

// ProjectNDC returns (ndcX, ndcY, inFront) for a world point p under the camera described by
// (basis b, eye, fov degrees, aspect = rectWidth/rectHeight). inFront is false when p is on
// or behind the camera plane (three.js ndc.z > 1 skip).
func ProjectNDC(p, eye vec3, b CamBasis, fovDeg, aspect float64) (ndcX, ndcY float64, inFront bool) {
	rel := p.Sub(eye)
	cx := rel.Dot(b.RefX)
	cy := rel.Dot(b.RefY)
	cz := rel.Dot(b.Pole) // +Z toward camera; a point in front has cz < 0
	if cz >= 0 {
		return 0, 0, false
	}
	tanHalf := math.Tan((fovDeg * math.Pi / 180) / 2)
	ndcX = cx / (-cz) / (tanHalf * aspect)
	ndcY = cy / (-cz) / tanHalf
	return ndcX, ndcY, true
}

// RayDirThroughNDC mirrors THREE.Raycaster.setFromCamera for a perspective camera: the world
// ray direction from the eye through NDC (nx, ny). camera-space dir = (nx*tanHalf*aspect,
// ny*tanHalf, -1), rotated into world by the basis (pole is +Z, so -1 along Z faces forward).
func RayDirThroughNDC(nx, ny float64, b CamBasis, fovDeg, aspect float64) vec3 {
	tanHalf := math.Tan((fovDeg * math.Pi / 180) / 2)
	camDir := b.RefX.Scale(nx * tanHalf * aspect).
		Add(b.RefY.Scale(ny * tanHalf)).
		Add(b.Pole.Scale(-1))
	return camDir.Normalize()
}
