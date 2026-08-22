package scenepersist

import (
	"github.com/dtauraso/wirefold/src/Polar/polar"
	"github.com/dtauraso/wirefold/src/Scene"
	"github.com/dtauraso/wirefold/src/Scene/viewstate"
)

func LoadSceneSphere(topologyPath string) (polar.SceneSphere, bool) {
	var s polar.SceneSphere
	read := func(name string, dst *float64) bool {
		return ReadIfExists(Scene.SceneValuePath(topologyPath, name), dst)
	}
	if !read("cx", &s.Center.X) || !read("cy", &s.Center.Y) ||
		!read("cz", &s.Center.Z) || !read("radius", &s.Radius) {
		return polar.SceneSphere{}, false
	}
	return s, true
}

func WriteSceneSphere(sceneRoot string, s polar.SceneSphere) error {
	for name, value := range map[string]float64{
		"cx": s.Center.X, "cy": s.Center.Y, "cz": s.Center.Z,
		"radius": s.Radius,
	} {
		if err := WriteAtomicIfChanged(Scene.SceneValuePath(sceneRoot, name), value); err != nil {
			return err
		}
	}
	return nil
}

func InstallSceneSphere(ui *viewstate.UIState, gs *Scene.GeomSeeds, topologyPath string) {
	if s, ok := LoadSceneSphere(topologyPath); ok {
		ui.SceneSphere = s
	} else {

		centers := gs.LoadTimeCenters()
		polarCenters := make(map[string]polar.Vec3, len(centers))
		for id, c := range centers {
			polarCenters[id] = polar.Vec3(c)
		}
		ui.SceneSphere = polar.ContentFitSceneSphere(polarCenters)

		if topologyPath != "" {
			_ = WriteSceneSphere(topologyPath, ui.SceneSphere)
		}
	}

	ui.EmitViewFrame(nil)
}
