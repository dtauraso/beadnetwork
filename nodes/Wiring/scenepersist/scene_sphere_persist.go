package scenepersist

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom"
	"github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	"github.com/dtauraso/wirefold/nodes/spatial"
	wire "github.com/dtauraso/wirefold/nodes/wire"

	T "github.com/dtauraso/wirefold/Trace"
)

type sceneSphereJSON struct {
	Center *[3]float64 `json:"center"`
	Radius *float64    `json:"radius"`
}

func LoadSceneSphere(topologyPath string) (geom.SceneSphere, bool) {
	var sj sceneSphereJSON
	jsonpersist.ReadJSONBestEffort(scenepaths.SphereFilePath(topologyPath), &sj)
	if sj.Center == nil || sj.Radius == nil {
		return geom.SceneSphere{}, false
	}
	return geom.SceneSphere{
		Center: spatial.Vec3{X: sj.Center[0], Y: sj.Center[1], Z: sj.Center[2]},
		Radius: *sj.Radius,
	}, true
}

func WriteSceneSphere(sphereJSONPath string, s geom.SceneSphere) error {
	center := [3]float64{s.Center.X, s.Center.Y, s.Center.Z}
	radius := s.Radius
	return jsonpersist.WriteJSONAtomic(sphereJSONPath, sceneSphereJSON{Center: &center, Radius: &radius})
}

func InstallSceneSphere(ui *viewstate.UIState, gs *geomseeds.GeomSeeds, topologyPath string) {
	if s, ok := LoadSceneSphere(topologyPath); ok {
		ui.SceneSphere = s
	} else {

		ui.SceneSphere = geom.ContentFitSceneSphere(gs.LoadTimeCenters())

		if topologyPath != "" {
			_ = WriteSceneSphere(scenepaths.SphereFilePath(topologyPath), ui.SceneSphere)
		}
	}

	ui.EmitViewFrame([]wire.RowEvent{{Kind: T.KindSceneSphere, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1}})
}
