package Camera

import "math"

func ProjectNDC(p, eye vec3, b CamBasis, fovDeg, aspect float64) (ndcX, ndcY float64, inFront bool) {
	rel := p.Sub(eye)
	cx := rel.Dot(b.RefX)
	cy := rel.Dot(b.RefY)
	cz := rel.Dot(b.Pole)
	if cz >= 0 {
		return 0, 0, false
	}
	tanHalf := math.Tan((fovDeg * math.Pi / 180) / 2)
	ndcX = cx / (-cz) / (tanHalf * aspect)
	ndcY = cy / (-cz) / tanHalf
	return ndcX, ndcY, true
}

func RayDirThroughNDC(nx, ny float64, b CamBasis, fovDeg, aspect float64) vec3 {
	tanHalf := math.Tan((fovDeg * math.Pi / 180) / 2)
	camDir := b.RefX.Scale(nx * tanHalf * aspect).
		Add(b.RefY.Scale(ny * tanHalf)).
		Add(b.Pole.Scale(-1))
	return camDir.Normalize()
}
