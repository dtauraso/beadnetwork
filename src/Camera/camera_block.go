package Camera

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/valuefile"
)

const BlockRelPath = "view/camera.bin"

var BlockValueNames = []string{
	"pivotX", "pivotY", "pivotZ",
	"r",
	"posPhi", "posTheta",
	"upPhi", "upTheta",
}

func BlockPath(sceneRoot string) string {
	return filepath.Join(sceneRoot, filepath.FromSlash(BlockRelPath))
}

func writeViewpointBlock(path string, v Viewpoint) error {
	w := valuefile.NewBlobWriter(path, BlockValueNames)
	w.Begin()
	w.F64("pivotX", v.Pivot.X)
	w.F64("pivotY", v.Pivot.Y)
	w.F64("pivotZ", v.Pivot.Z)
	w.F64("r", v.R)
	w.F64("posPhi", v.Pos.Phi)
	w.F64("posTheta", v.Pos.Theta)
	w.F64("upPhi", v.Up.Phi)
	w.F64("upTheta", v.Up.Theta)
	return w.Flush()
}

func readViewpointBlock(path string) (Viewpoint, bool) {
	var v Viewpoint
	r, ok := valuefile.ReadBlob(path, BlockValueNames)
	if !ok {
		return v, false
	}
	get := func(name string, dst *float64) bool {
		got, ok := r.F64(name)
		if !ok {
			return false
		}
		*dst = got
		return true
	}
	if !get("pivotX", &v.Pivot.X) || !get("pivotY", &v.Pivot.Y) || !get("pivotZ", &v.Pivot.Z) {
		return Viewpoint{}, false
	}
	if !get("r", &v.R) || !get("posPhi", &v.Pos.Phi) || !get("posTheta", &v.Pos.Theta) {
		return Viewpoint{}, false
	}
	if !get("upPhi", &v.Up.Phi) || !get("upTheta", &v.Up.Theta) {
		return Viewpoint{}, false
	}
	return v, true
}
