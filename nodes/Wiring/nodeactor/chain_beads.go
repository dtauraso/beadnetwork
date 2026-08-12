package nodeactor

import (
	"os"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/beadindex"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

var chainAimTraceEnabled = os.Getenv("WIREFOLD_CHAIN_AIM_TRACE") == "1"

func (m *NodeGeometry) chainBeads() (ox, oy, oz []float32, lit []uint8, litVal []int32, breadcrumbs []wire.RowEvent) {
	if len(m.outs.outTargets) == 0 {
		return nil, nil, nil, nil, nil, nil
	}

	var tick int64
	if len(m.outs.outWires) > 0 {
		tick = m.clocks.clk.Tick()
	}
	selfTorusR := nodegeom.NodeTorusOuterR(m.geom.Kind)

	selfCenter := nodegeom.NodeWorldPos(m.geom)
	for _, to := range m.outs.outTargets {

		targetCenter, haveTargetCenter := m.topo.partnerCenters[to]
		if !haveTargetCenter {
			continue
		}
		edgeOX, edgeOY, edgeOZ, edgeLit, edgeLitVal, breadcrumb, ok := m.chainBeadsForTarget(to, tick, selfTorusR, selfCenter, targetCenter)
		if !ok {
			continue
		}
		ox = append(ox, edgeOX...)
		oy = append(oy, edgeOY...)
		oz = append(oz, edgeOZ...)
		lit = append(lit, edgeLit...)
		litVal = append(litVal, edgeLitVal...)
		if breadcrumb != nil {
			breadcrumbs = append(breadcrumbs, *breadcrumb)
		}
	}
	return ox, oy, oz, lit, litVal, breadcrumbs
}

func (m *NodeGeometry) chainBeadsForTarget(to string, tick int64, selfTorusR float64, selfCenter, targetCenter vec3) (ox, oy, oz []float32, lit []uint8, litVal []int32, breadcrumb *wire.RowEvent, ok bool) {

	dist, liveDir, count, geomOK := beadindex.ChainEdgeGeometry(selfCenter, targetCenter, selfTorusR, m.geom.Kind, m.topo.neighborKinds[to])
	if !geomOK {
		return nil, nil, nil, nil, nil, nil, false
	}

	m.publishStepCount(to, count)

	pulses := m.gatherPulses(to, tick)

	breadcrumb = m.chainAimBreadcrumb(to, count, dist, liveDir)

	step := lattice.BeadStepR
	base := selfTorusR + lattice.BeadTorusOuterR
	offsetAt := func(i int) float64 {
		return beadindex.BeadPlacementOffset(base, step, i)
	}

	aimUnit := liveDir

	var chainSep vec3
	if m.topo.mutualTargets[to] {
		if off, sepOK := nodegeom.ParallelChainOffset(m.id, to, selfCenter, targetCenter, m.geom.SceneCenter); sepOK {
			chainSep = off
		}
	}

	var actorChain *edgeBeadChain
	if m.beads.beadTickFn != nil {
		actorChain = m.reconcileBeadChain(to, count, offsetAt, aimUnit)
	}

	var resolved []vec3
	var resolvedValid []bool
	if actorChain != nil {
		resolved = make([]vec3, len(actorChain.last))
		resolvedValid = actorChain.valid
		for i, s := range actorChain.last {
			resolved[i] = s.Position
		}
	}

	ox, oy, oz, lit, litVal = beadindex.ChainBeadRows(liveDir, chainSep, base, step, count, resolved, resolvedValid, pulses)
	return ox, oy, oz, lit, litVal, breadcrumb, true
}

func (m *NodeGeometry) publishStepCount(to string, count int) {
	for i, wt := range m.outs.outWireTargets {
		if wt != to {
			continue
		}
		if i < len(m.outs.outWireOuts) && m.outs.outWireOuts[i] != nil {
			m.outs.outWireOuts[i].PublishSteps(count)
		}
		if i < len(m.outs.outStepsIn) && m.outs.outStepsIn[i] != nil {
			m.outs.outStepsIn[i](count)
		}
	}
}

func (m *NodeGeometry) gatherPulses(to string, tick int64) []beadindex.Pulse {
	var pulses []beadindex.Pulse
	for i, wt := range m.outs.outWireTargets {
		if wt != to || m.outs.outWires[i] == nil {
			continue
		}
		for _, p := range m.outs.outWires[i].LiveBeadFractions(tick) {
			if p.T < 0 || p.T >= 1 || p.Steps <= 0 {
				continue
			}

			pulses = append(pulses, beadindex.Pulse{T: p.T, Steps: p.Steps, Val: int32(p.Val)})
		}
	}
	return pulses
}

func (m *NodeGeometry) chainAimBreadcrumb(to string, count int, dist float64, liveDir vec3) *wire.RowEvent {
	if m.tr == nil || !chainAimTraceEnabled {
		return nil
	}
	targetRow := int32(-1)
	if m.topo.nodeRowFor != nil {
		if r, ok := m.topo.nodeRowFor(to); ok {
			targetRow = r
		}
	}
	value := beadindex.ChainAimBreadcrumbText(to, count, dist, liveDir)
	m.tr.Breadcrumb("chain-aim", m.id, to, value)
	return &wire.RowEvent{
		Kind: T.KindBreadcrumb, Label: T.BreadcrumbChainAim, Debug: 1,
		NodeRow: m.stream.nodeRow, PortRow: -1, TargetRow: targetRow, TargetPortRow: -1,
		EdgeRow: -1, Slot: -1, Text: value,
	}
}
