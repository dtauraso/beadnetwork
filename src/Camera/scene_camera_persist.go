package Camera

import ()

type ViewpointPersister struct {
	Path string
}

func (p *ViewpointPersister) Schedule(v Viewpoint) {
	if p == nil || p.Path == "" {
		return
	}
	if err := writeViewpointBlock(p.Path, v); err != nil {
		LogPersistErr("scene_camera_persist", p.Path, err)
	}
}

func WriteSceneCamera(path string, v Viewpoint) error {
	return writeViewpointBlock(path, v)
}

func ReadSceneCamera(path string) (Viewpoint, bool) {
	return readViewpointBlock(path)
}
