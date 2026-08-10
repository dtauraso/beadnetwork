package Wiring

import "math"

// vecClose is a shared test-only tolerance comparator for vec3, used by several camera/
// gesture tests. Mirrors geom's own private vecClose (nodes/Wiring/geom/gesture_camera_test.go)
// — kept here too since Wiring's own tests reach it across package boundaries where geom's
// copy, being unexported, does not resolve.
func vecClose(a, b vec3, tol float64) bool {
	return math.Abs(a.X-b.X) < tol && math.Abs(a.Y-b.Y) < tol && math.Abs(a.Z-b.Z) < tol
}
