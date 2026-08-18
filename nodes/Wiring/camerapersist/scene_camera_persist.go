package camerapersist

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/camera"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

const (
	FilePivotX   = "pivot-x.json"
	FilePivotY   = "pivot-y.json"
	FilePivotZ   = "pivot-z.json"
	FileR        = "r.json"
	FilePosPhi   = "pos-phi.json"
	FilePosTheta = "pos-theta.json"
	FileUpPhi    = "up-phi.json"
	FileUpTheta  = "up-theta.json"
)

type ViewpointPersister struct {
	Dir string
}

func (p *ViewpointPersister) Schedule(v camera.Viewpoint) {
	if p == nil || p.Dir == "" {
		return
	}
	if err := WriteSceneCamera(p.Dir, v); err != nil {
		jsonpersist.LogPersistErr("scene_camera_persist", p.Dir, err)
	}
}

func WriteSceneCamera(dir string, v camera.Viewpoint) error {
	for name, value := range map[string]float64{
		FilePivotX: v.Pivot.X, FilePivotY: v.Pivot.Y, FilePivotZ: v.Pivot.Z,
		FileR:        v.R,
		FilePosPhi:   v.Pos.Phi,
		FilePosTheta: v.Pos.Theta,
		FileUpPhi:    v.Up.Phi,
		FileUpTheta:  v.Up.Theta,
	} {
		if err := jsonpersist.WriteJSONAtomic(filepath.Join(dir, name), value); err != nil {
			return err
		}
	}
	return nil
}

func ReadSceneCamera(dir string) (v camera.Viewpoint, ok bool) {
	read := func(name string, dst *float64) bool {
		return jsonpersist.ReadJSONIfExists(filepath.Join(dir, name), dst)
	}
	if !read(FilePivotX, &v.Pivot.X) || !read(FilePivotY, &v.Pivot.Y) || !read(FilePivotZ, &v.Pivot.Z) {
		return camera.Viewpoint{}, false
	}
	if !read(FileR, &v.R) {
		return camera.Viewpoint{}, false
	}
	if !read(FilePosPhi, &v.Pos.Phi) || !read(FilePosTheta, &v.Pos.Theta) {
		return camera.Viewpoint{}, false
	}
	if !read(FileUpPhi, &v.Up.Phi) || !read(FileUpTheta, &v.Up.Theta) {
		return camera.Viewpoint{}, false
	}
	return v, true
}
