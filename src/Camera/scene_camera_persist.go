package Camera

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/src/valuefile"
)

const (
	FilePivotX   = "pivot-x.bin"
	FilePivotY   = "pivot-y.bin"
	FilePivotZ   = "pivot-z.bin"
	FileR        = "r.bin"
	FilePosPhi   = "pos-phi.bin"
	FilePosTheta = "pos-theta.bin"
	FileUpPhi    = "up-phi.bin"
	FileUpTheta  = "up-theta.bin"
)

type ViewpointPersister struct {
	Dir string
}

func (p *ViewpointPersister) Schedule(v Viewpoint) {
	if p == nil || p.Dir == "" {
		return
	}
	for name, value := range viewpointValues(v) {
		if err := valuefile.WriteAtomicIfChanged(filepath.Join(p.Dir, name), value); err != nil {
			valuefile.LogPersistErr("scene_camera_persist", p.Dir, err)
			return
		}
	}
}

func viewpointValues(v Viewpoint) map[string]float64 {
	return map[string]float64{
		FilePivotX: v.Pivot.X, FilePivotY: v.Pivot.Y, FilePivotZ: v.Pivot.Z,
		FileR:        v.R,
		FilePosPhi:   v.Pos.Phi,
		FilePosTheta: v.Pos.Theta,
		FileUpPhi:    v.Up.Phi,
		FileUpTheta:  v.Up.Theta,
	}
}

func WriteSceneCamera(dir string, v Viewpoint) error {
	for name, value := range viewpointValues(v) {
		if err := valuefile.WriteAtomic(filepath.Join(dir, name), value); err != nil {
			return err
		}
	}
	return nil
}

func ReadSceneCamera(dir string) (v Viewpoint, ok bool) {
	read := func(name string, dst *float64) bool {
		return valuefile.ReadIfExists(filepath.Join(dir, name), dst)
	}
	if !read(FilePivotX, &v.Pivot.X) || !read(FilePivotY, &v.Pivot.Y) || !read(FilePivotZ, &v.Pivot.Z) {
		return Viewpoint{}, false
	}
	if !read(FileR, &v.R) {
		return Viewpoint{}, false
	}
	if !read(FilePosPhi, &v.Pos.Phi) || !read(FilePosTheta, &v.Pos.Theta) {
		return Viewpoint{}, false
	}
	if !read(FileUpPhi, &v.Up.Phi) || !read(FileUpTheta, &v.Up.Theta) {
		return Viewpoint{}, false
	}
	return v, true
}
