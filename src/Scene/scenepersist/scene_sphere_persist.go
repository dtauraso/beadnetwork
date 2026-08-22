package scenepersist

import (
	"github.com/dtauraso/wirefold/src/Polar/polar"
	"github.com/dtauraso/wirefold/src/Scene"
	"github.com/dtauraso/wirefold/src/Scene/viewstate"
	"github.com/dtauraso/wirefold/src/runtopology/geomseeds"
	"github.com/dtauraso/wirefold/src/valuefile"
)

func LoadSceneSphere(topologyPath string) (polar.SceneSphere, bool) {
	var s polar.SceneSphere
	read := func(name string, dst *float64) bool {
		return valuefile.ReadIfExists(Scene.SceneValuePath(topologyPath, name), dst)
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
		if err := valuefile.WriteAtomicIfChanged(Scene.SceneValuePath(sceneRoot, name), value); err != nil {
			return err
		}
	}
	return nil
}

func InstallSceneSphere(ui *viewstate.UIState, gs *geomseeds.GeomSeeds, topologyPath string) {
	if s, ok := LoadSceneSphere(topologyPath); ok {
		ui.SceneSphere = s
	} else {

		ui.SceneSphere = polar.ContentFitSceneSphere(gs.LoadTimeCenters())

		if topologyPath != "" {
			_ = WriteSceneSphere(topologyPath, ui.SceneSphere)
		}
	}

	ui.EmitViewFrame(nil)
}
