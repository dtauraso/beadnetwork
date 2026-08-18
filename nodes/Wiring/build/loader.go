package build

import (
	"context"
	"github.com/dtauraso/wirefold/tools/topology-vscode/SliderPanel"

	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/kindreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/nodes/clock"
	"github.com/dtauraso/wirefold/nodes/nodeapi"

	T "github.com/dtauraso/wirefold/tools/topology-vscode/Trace"
)

func LoadTopology(ctx context.Context, jsonPath string, tr *T.Trace, clk clock.Clock) ([]nodeapi.Node, inputcodec.SlotRegistry, *dispatch.MoveDispatch, SliderPanel.Sinks, error) {
	kindreg.BuildRegistry()
	spec, err := loadspec.ParseSpec(jsonPath)
	if err != nil {
		return nil, nil, nil, SliderPanel.Sinks{}, err
	}

	kindPorts := make(map[string][]portwiring.PortSpec, len(kindreg.Registry))
	for kind, bind := range kindreg.Registry {
		kindPorts[kind] = bind.Ports
	}
	if err := loadspec.ValidateSpec(&spec, kindPorts); err != nil {
		return nil, nil, nil, SliderPanel.Sinks{}, err
	}

	sphere, hasScene := scenepersist.LoadSceneSphere(jsonPath)
	return buildFromSpec(ctx, spec, tr, clk, sphere, hasScene, jsonPath)
}
