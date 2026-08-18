package scenepersist

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/geomseeds"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	"github.com/dtauraso/wirefold/nodes/spatial"

	T "github.com/dtauraso/wirefold/tools/topology-vscode/Trace"
)

type sceneSphereJSON struct {
	Center *[3]float64 `json:"center"`
	Radius *float64    `json:"radius"`
}

func LoadSceneSphere(topologyPath string) (polar.SceneSphere, bool) {
	var sj sceneSphereJSON
	jsonpersist.ReadJSONBestEffort(scenepaths.SphereFilePath(topologyPath), &sj)
	if sj.Center == nil || sj.Radius == nil {
		return polar.SceneSphere{}, false
	}
	return polar.SceneSphere{
		Center: spatial.Vec3{X: sj.Center[0], Y: sj.Center[1], Z: sj.Center[2]},
		Radius: *sj.Radius,
	}, true
}

func WriteSceneSphere(sphereJSONPath string, s polar.SceneSphere) error {
	center := [3]float64{s.Center.X, s.Center.Y, s.Center.Z}
	radius := s.Radius
	return jsonpersist.WriteJSONAtomic(sphereJSONPath, sceneSphereJSON{Center: &center, Radius: &radius})
}

func InstallSceneSphere(ui *viewstate.UIState, gs *geomseeds.GeomSeeds, topologyPath string) {
	if s, ok := LoadSceneSphere(topologyPath); ok {
		ui.SceneSphere = s
	} else {

		ui.SceneSphere = polar.ContentFitSceneSphere(gs.LoadTimeCenters())

		if topologyPath != "" {
			_ = WriteSceneSphere(scenepaths.SphereFilePath(topologyPath), ui.SceneSphere)
		}
	}

	ui.EmitViewFrame([]rowevent.RowEvent{{Kind: T.KindSceneSphere, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1}})
}
