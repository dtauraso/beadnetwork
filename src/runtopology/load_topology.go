package runtopology

import (
	"context"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"

	"github.com/dtauraso/wirefold/src/Input/dispatch"
	"github.com/dtauraso/wirefold/src/Input/inputcodec"
	"github.com/dtauraso/wirefold/src/NodeKinds/kindreg"
	"github.com/dtauraso/wirefold/src/runtopology/loadspec"
	"github.com/dtauraso/wirefold/src/NodeKinds/portwiring"
	"github.com/dtauraso/wirefold/src/Scene/scenepersist"
	"github.com/dtauraso/wirefold/src/Clock"
	"github.com/dtauraso/wirefold/src/NodeKinds/nodeapi"
)

func LoadTopology(ctx context.Context, jsonPath string, clk clock.Clock) ([]nodeapi.Node, inputcodec.SlotRegistry, *dispatch.MoveDispatch, SliderPanel.Sinks, error) {
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
	return buildFromSpec(ctx, spec, clk, sphere, hasScene, jsonPath)
}
