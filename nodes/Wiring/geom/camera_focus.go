package geom

// camera_focus.go — scene geometry for gesture pivots (mirrors geometry-helpers.ts
// contentSphere + interaction-handlers.ts regionFocus), split from camera_angles.go by
// concern (see that file's header for the shared quarantine rationale).

import "math"

const GestureFocusMin = 10.0 // FOCUS_MIN — keep the RegionFocus pivot off the camera
const GestureZoomBase = 1.01 // ZOOM_BASE — per-scroll-unit dolly factor

// RotSmoothAlpha is the EMA factor for the averaging ("fat") cursor that drives rotation:
// LOWER = fatter/blurrier/smoother, HIGHER = snappier/closer to raw. Range (0,1].
const RotSmoothAlpha = 0.35

// FocusAhead returns the orbit center for rotate: a point on the view-center ray at the
// forward-depth of the node the camera is MOST POINTED AT (smallest angle from the view axis,
// in front). Because the point lies on the view axis, orbiting it does NOT re-aim the camera —
// the look direction is unchanged — yet the orbit depth tracks whatever content you have flown
// to and centered (fly to node 10, rotate spins around node 10). Falls back to a fixed distance
// ahead when there is no node in front.
func FocusAhead(v Viewpoint, centers map[string]vec3) vec3 {
	eye := EyeOf(v)
	forward := AnglesToWorldOffset(1, v.Pos.Theta, v.Pos.Phi).Scale(-1) // -pole, unit
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
		if cosAng <= 0 { // behind the camera
			continue
		}
		if cosAng > bestCos { // more centered on the view axis
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

// ContentSphereOf mirrors geometry-helpers.ts contentSphere over the given node centers:
// center = bbox midpoint, radius = max(center-distance)*1.1 (min 1). Empty → (origin, 100).
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

// RegionFocus mirrors interaction-handlers.ts regionFocus: the center of the node depth slab
// straight ahead of the camera. forward = -pole (camera looks along -Z); depth of each node
// = forward · (p - eye); pivot = eye + forward * max((zNear+zFar)/2, FOCUS_MIN). Falls back
// to eye + forward*FOCUS_MIN when there are no finite node depths.
func RegionFocus(v Viewpoint, centers map[string]vec3) vec3 {
	eye := EyeOf(v)
	forward := AnglesToWorldOffset(1, v.Pos.Theta, v.Pos.Phi).Scale(-1) // -pole, unit
	// Pivot on the view axis at the depth of the NEAREST node (smallest forward-depth), not the
	// whole-scene depth MIDPOINT. The midpoint sat between the near node you zoomed into and the
	// far ones, so rotate/pan operated around a distant point and swung/overshot from up close.
	// Using the nearest depth puts the pivot on what you dollied toward while staying on the
	// screen-center ray (deterministic — no node-tie ambiguity). Falls back straight ahead when
	// there are no nodes.
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
