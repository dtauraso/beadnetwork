package Camera

import (
	"math"
)

const GestureFocusMin = 10.0
const GestureZoomBase = 1.01

const RotationUnitEyeRadii = 4.0

const RotationInsideBoost = 4.0

const RotationMaxScale = 240.0

func RotationScale(v Viewpoint, sphereCentre Vec3, sphereRadius float64) float64 {
	if sphereRadius <= 0 {
		return 1
	}
	eyeDist := EyeOf(v).Sub(sphereCentre).Length()
	if eyeDist < GestureFocusMin {
		eyeDist = GestureFocusMin
	}
	scale := RotationUnitEyeRadii * sphereRadius / eyeDist
	if eyeDist < sphereRadius {
		scale *= RotationInsideBoost
	}
	if scale > RotationMaxScale {
		scale = RotationMaxScale
	}
	return scale
}

func FocusAhead(v Viewpoint, centers map[string]Vec3) Vec3 {
	eye := EyeOf(v)
	forward := AnglesToWorldOffset(1, v.Pos.Phi, v.Pos.Theta).Scale(-1)
	bestCos := -2.0
	depth := 0.0
	found := false
	for _, p := range centers {
		d := p.Sub(eye)
		dl := d.Length()
		if dl < 1e-9 {
			continue
		}
		cosAng := forward.Dot(d) / dl
		if cosAng <= 0 {
			continue
		}
		if cosAng > bestCos {
			bestCos = cosAng
			depth = forward.Dot(d)
			found = true
		}
	}
	if !found {
		return eye.Add(forward.Scale(GestureFocusMin))
	}
	return eye.Add(forward.Scale(math.Max(depth, GestureFocusMin)))
}

func RegionFocus(v Viewpoint, centers map[string]Vec3) Vec3 {
	eye := EyeOf(v)
	forward := AnglesToWorldOffset(1, v.Pos.Phi, v.Pos.Theta).Scale(-1)

	zNear := math.Inf(1)
	for _, p := range centers {
		depth := forward.Dot(p.Sub(eye))
		if math.IsNaN(depth) || math.IsInf(depth, 0) {
			continue
		}
		zNear = math.Min(zNear, depth)
	}
	if math.IsInf(zNear, 1) {
		return eye.Add(forward.Scale(GestureFocusMin))
	}
	return eye.Add(forward.Scale(math.Max(zNear, GestureFocusMin)))
}
