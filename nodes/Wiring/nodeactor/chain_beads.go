package nodeactor

import (
	"os"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/beadindex"
	"github.com/dtauraso/wirefold/nodes/Wiring/edgegeom"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodegeom"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	lattice "github.com/dtauraso/wirefold/nodes/wire/lattice"
)

var chainAimTraceEnabled = os.Getenv("WIREFOLD_CHAIN_AIM_TRACE") == "1"

func (m *NodeGeometry) chainBeads() (ox, oy, oz []float32, lit []uint8, litVal []int32, breadcrumbs []rowevent.RowEvent) {
	if len(m.outTargets) == 0 {
		return nil, nil, nil, nil, nil, nil
	}

	m.drainPulses()

	selfTorusR := nodegeom.NodeTorusOuterR(m.geom.Kind)

	selfCenter := nodegeom.NodeWorldPos(m.geom)
	counts := make(map[string]int, len(m.outTargets))
	for _, to := range m.outTargets {

		// The stored outward path IS the animation path — the direction and
		// length the beads travel come from it, not from a direction
		// recomputed out of two node centres on every pass.
		path, havePath := m.topo.PathTo(to)
		if !havePath {
			continue
		}
		edgeOX, edgeOY, edgeOZ, edgeLit, edgeLitVal, breadcrumb, ok := m.chainBeadsForTarget(to, selfTorusR, selfCenter, selfCenter.Add(path), counts)
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
	m.sendStepCounts(counts)
	return ox, oy, oz, lit, litVal, breadcrumbs
}

// outPoles is one unit direction per OUTGOING neighbour, in outTargets order:
// the stored path vector's direction, which the editor draws as that edge's
// +y pole. n outgoing neighbours, n poles.
func (m *NodeGeometry) outPoles() (dx, dy, dz []float32) {
	for _, to := range m.outTargets {
		path, ok := m.topo.PathTo(to)
		if !ok {
			continue
		}
		unit := path.Normalize()
		dx = append(dx, float32(unit.X))
		dy = append(dy, float32(unit.Y))
		dz = append(dz, float32(unit.Z))
	}
	return dx, dy, dz
}

// drainPulses applies the most recently received gathered-pulses message
// from the animation peer, non-blocking. Between animation cycles geometry
// keeps rendering with the last pulses it received.
func (m *NodeGeometry) drainPulses() {
	for {
		select {
		case p := <-m.pulseIn:
			m.lastPulses = p
		default:
			return
		}
	}
}

// sendStepCounts hands this pass's step counts to the animation peer,
// non-blocking — the geometry side's answer to "how many steps to the
// target", sent once per target set rather than per target.
func (m *NodeGeometry) sendStepCounts(counts map[string]int) {
	if len(counts) == 0 {
		return
	}
	select {
	case m.stepOut <- counts:
	default:
	}
}

func (m *NodeGeometry) chainBeadsForTarget(to string, selfTorusR float64, selfCenter, targetCenter vec3, counts map[string]int) (ox, oy, oz []float32, lit []uint8, litVal []int32, breadcrumb *rowevent.RowEvent, ok bool) {

	dist, liveDir, count, geomOK := beadindex.ChainEdgeGeometry(selfCenter, targetCenter, selfTorusR, m.geom.Kind, m.topo.NeighborKind(to))
	if !geomOK {
		return nil, nil, nil, nil, nil, nil, false
	}

	counts[to] = count

	pulses := m.lastPulses[to]

	breadcrumb = m.chainAimBreadcrumb(to, count, dist, liveDir)

	step := lattice.BeadStepR
	base := selfTorusR + lattice.BeadTorusOuterR
	offsetAt := func(i int) float64 {
		return beadindex.BeadPlacementOffset(base, step, i)
	}

	aimUnit := liveDir

	var chainSep vec3
	if m.topo.IsMutualTarget(to) {
		if off, sepOK := edgegeom.ParallelChainOffset(m.id, to, selfCenter, targetCenter, m.geom.SceneCenter); sepOK {
			chainSep = off
		}
	}

	actorChain := m.beads.ReconcileBeadChain(to, count, offsetAt, aimUnit)
	resolved, resolvedValid := actorChain.Resolved()

	ox, oy, oz, lit, litVal = beadindex.ChainBeadRows(liveDir, chainSep, base, step, count, resolved, resolvedValid, pulses)
	return ox, oy, oz, lit, litVal, breadcrumb, true
}

func (m *NodeGeometry) chainAimBreadcrumb(to string, count int, dist float64, liveDir vec3) *rowevent.RowEvent {
	if m.tr == nil || !chainAimTraceEnabled {
		return nil
	}
	targetRow := int32(-1)
	if r, ok := m.topo.NodeRowFor(to); ok {
		targetRow = r
	}
	value := beadindex.ChainAimBreadcrumbText(to, count, dist, liveDir)
	m.tr.Breadcrumb("chain-aim", m.id, to, value)
	return &rowevent.RowEvent{
		Kind: T.KindBreadcrumb, Label: T.BreadcrumbChainAim, Debug: 1,
		NodeRow: m.stream.NodeRow(), PortRow: -1, TargetRow: targetRow, TargetPortRow: -1,
		EdgeRow: -1, Slot: -1, Text: value,
	}
}
