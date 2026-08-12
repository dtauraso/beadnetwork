package camerapersist

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

type PolarCamera struct {
	Pivot *[3]float64 `json:"pivot"`
	R     *float64    `json:"r"`
	Pos   *[2]float64 `json:"pos"`
	Up    *[2]float64 `json:"up"`
}

type ViewpointPersister struct {
	Path string
}

func (p *ViewpointPersister) Schedule(v geom.Viewpoint) {
	if p == nil || p.Path == "" {
		return
	}
	cam := ViewpointToPolar(v)
	if err := WriteSceneCameraPolar(p.Path, cam); err != nil {
		jsonpersist.LogPersistErr("scene_camera_persist", p.Path, err)
		return
	}
}

func ViewpointToPolar(v geom.Viewpoint) *PolarCamera {
	pivot := [3]float64{v.Pivot.X, v.Pivot.Y, v.Pivot.Z}
	r := v.R
	pos := [2]float64{v.Pos.Theta, v.Pos.Phi}
	up := [2]float64{v.Up.Theta, v.Up.Phi}
	return &PolarCamera{Pivot: &pivot, R: &r, Pos: &pos, Up: &up}
}

func WriteSceneCameraPolar(path string, cam *PolarCamera) error {
	return jsonpersist.WriteJSONAtomic(path, cam)
}
