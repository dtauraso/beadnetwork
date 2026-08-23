package Startup

import (
	"github.com/dtauraso/wirefold/Categories/Polar/polar"
	SceneBuf "github.com/dtauraso/wirefold/Categories/Scene"
	"github.com/dtauraso/wirefold/Categories/Scene/viewpersist"
	"github.com/dtauraso/wirefold/Categories/Scene/viewstate"
)

func LoadSceneSphere(topologyPath string) (polar.SceneSphere, bool) {
	var s polar.SceneSphere
	read := func(name string, dst *float64) bool {
		return ReadIfExists(SceneBuf.SceneValuePath(topologyPath, name), dst)
	}
	if !read("cx", &s.Center.X) || !read("cy", &s.Center.Y) ||
		!read("cz", &s.Center.Z) || !read("radius", &s.Radius) {
		return polar.SceneSphere{}, false
	}
	return s, true
}

func InstallSceneSphere(ui *viewstate.UIState, gs *SceneBuf.GeomSeeds, topologyPath string) {
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
			_ = viewpersist.WriteSceneSphere(topologyPath, ui.SceneSphere)
		}
	}

	ui.EmitViewFrame(nil)
}
