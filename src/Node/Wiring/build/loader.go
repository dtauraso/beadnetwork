package build

import (
	"context"
	"github.com/dtauraso/wirefold/src/SliderPanel"

	"github.com/dtauraso/wirefold/src/Node/Wiring/dispatch"
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/kindreg"
	"github.com/dtauraso/wirefold/src/Node/Wiring/loadspec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/portwiring"
	"github.com/dtauraso/wirefold/src/Node/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/src/Node/clock"
	"github.com/dtauraso/wirefold/src/Node/nodeapi"

	T "github.com/dtauraso/wirefold/src/Trace"
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
