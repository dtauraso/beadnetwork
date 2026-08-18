package scenepersist

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

const (
	FileSphereCX     = "cx.json"
	FileSphereCY     = "cy.json"
	FileSphereCZ     = "cz.json"
	FileSphereRadius = "radius.json"
)

func LoadSceneSphere(topologyPath string) (polar.SceneSphere, bool) {
	dir := scenepaths.SphereDirPath(topologyPath)
	var s polar.SceneSphere
	read := func(name string, dst *float64) bool {
		return jsonpersist.ReadJSONIfExists(filepath.Join(dir, name), dst)
	}
	if !read(FileSphereCX, &s.Center.X) || !read(FileSphereCY, &s.Center.Y) ||
		!read(FileSphereCZ, &s.Center.Z) || !read(FileSphereRadius, &s.Radius) {
		return polar.SceneSphere{}, false
	}
	return s, true
}

func WriteSceneSphere(sphereDir string, s polar.SceneSphere) error {
	for name, value := range map[string]float64{
		FileSphereCX: s.Center.X, FileSphereCY: s.Center.Y, FileSphereCZ: s.Center.Z,
		FileSphereRadius: s.Radius,
	} {
		if err := jsonpersist.WriteJSONAtomicIfChanged(filepath.Join(sphereDir, name), value); err != nil {
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
			_ = WriteSceneSphere(scenepaths.SphereDirPath(topologyPath), ui.SceneSphere)
		}
	}

	ui.EmitViewFrame(nil)
}
