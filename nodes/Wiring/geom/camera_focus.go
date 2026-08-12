package geom

import "math"

const GestureFocusMin = 10.0
const GestureZoomBase = 1.01

const RotSmoothAlpha = 0.35

func FocusAhead(v Viewpoint, centers map[string]vec3) vec3 {
	eye := EyeOf(v)
	forward := AnglesToWorldOffset(1, v.Pos.Theta, v.Pos.Phi).Scale(-1)
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

func ContentSphereOf(centers map[string]vec3) (center vec3, radius float64) {
	if len(centers) == 0 {
		return vec3{}, 100
	}
	min := vec3{X: math.Inf(1), Y: math.Inf(1), Z: math.Inf(1)}
	max := vec3{X: math.Inf(-1), Y: math.Inf(-1), Z: math.Inf(-1)}
	for _, p := range centers {
		if math.IsInf(p.X, 0) || math.IsNaN(p.X) {
			continue
		}
		min.X, max.X = math.Min(min.X, p.X), math.Max(max.X, p.X)
		min.Y, max.Y = math.Min(min.Y, p.Y), math.Max(max.Y, p.Y)
		min.Z, max.Z = math.Min(min.Z, p.Z), math.Max(max.Z, p.Z)
	}
	center = min.Add(max).Scale(0.5)
	r := 0.0
	for _, p := range centers {
		r = math.Max(r, p.Sub(center).Length())
	}
	return center, math.Max(r*1.1, 1)
}

func RegionFocus(v Viewpoint, centers map[string]vec3) vec3 {
	eye := EyeOf(v)
	forward := AnglesToWorldOffset(1, v.Pos.Theta, v.Pos.Phi).Scale(-1)

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
