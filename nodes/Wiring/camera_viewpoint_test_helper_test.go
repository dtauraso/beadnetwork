package Wiring

// camera_viewpoint_test_helper_test.go — a test-only duplicate of
// nodes/Wiring/scenecamera's LoadSceneViewpoint, kept here for the same reason
// vec_close_test.go's own vecClose duplicate is: Wiring's own in-package tests reach
// this across a package boundary where scenecamera's copy would create a real import
// cycle (scenecamera already imports Wiring for *Wiring.MoveDispatch, so a same-package
// Wiring test cannot import scenecamera). No production code calls this; it exists only
// so an in-package test can assert what landed in camera.json without duplicating the
// completeness rule inline at every call site.

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/camerapersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
)

// loadSceneViewpoint mirrors scenecamera.LoadSceneViewpoint exactly (see that function's
// own doc comment for the field mapping and the "require every field" rule).
func loadSceneViewpoint(topologyPath string) (pivot vec3, r float64, pos, up geom.Dir, ok bool) {
	var cp camerapersist.PolarCamera
	jsonpersist.ReadJSONBestEffort(scenepaths.CameraFilePath(topologyPath), &cp)
	if cp.Pivot == nil || cp.R == nil || cp.Pos == nil || cp.Up == nil {
		return vec3{}, 0, geom.Dir{}, geom.Dir{}, false
	}
	pivot = vec3{X: cp.Pivot[0], Y: cp.Pivot[1], Z: cp.Pivot[2]}
	r = *cp.R
	pos = geom.Dir{Theta: cp.Pos[0], Phi: cp.Pos[1]}
	up = geom.Dir{Theta: cp.Up[0], Phi: cp.Up[1]}
	return pivot, r, pos, up, true
}
