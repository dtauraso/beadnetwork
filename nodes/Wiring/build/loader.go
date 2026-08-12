package build

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepersist"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"

	T "github.com/dtauraso/wirefold/Trace"
)

func LoadTopology(ctx context.Context, jsonPath string, tr *T.Trace, clk clock.Clock) ([]wire.Node, inputcodec.SlotRegistry, *dispatch.MoveDispatch, []chan float64, error) {
	kindapi.BuildRegistry()
	spec, err := loadspec.ParseSpec(jsonPath)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	kindPorts := make(map[string][]portwiring.PortSpec, len(kindapi.Registry))
	for kind, bind := range kindapi.Registry {
		kindPorts[kind] = bind.Ports
	}
	if err := loadspec.ValidateSpec(&spec, kindPorts); err != nil {
		return nil, nil, nil, nil, err
	}

	sphere, hasScene := scenepersist.LoadSceneSphere(jsonPath)
	return buildFromSpec(ctx, spec, tr, clk, sphere, hasScene, jsonPath)
}
