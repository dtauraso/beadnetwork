package Drag

import (

	"github.com/dtauraso/beadnetwork/Categories/Vector/polar"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Camera"
)

func (g *GestureState) BeginSphereRotation(vp Camera.Viewpoint, sphere polar.SceneSphere, heldCenters func() map[string]Vec3, ev RawInputMsg) {
	centers := heldCenters()
	camCenters := make(map[string]Camera.Vec3, len(centers))
	for id, c := range centers {
		camCenters[id] = Camera.Vec3(c)
	}

	pivot := Camera.FocusAhead(vp, camCenters)
	g.RotPivot = Vec3(pivot)
}
