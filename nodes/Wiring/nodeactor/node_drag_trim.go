package nodeactor

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
)

const SharedLengthKind = "Input"

func (m *NodeGeometry) TrimOwnDrag(delta polar.Polar) polar.Polar {
	if m.SelfKind() == SharedLengthKind {
		delta = polar.SnapDeltaTheta(delta)
	}
	delta = m.trimToOrbitRule(delta)
	return m.trimToEqualOutLengths(delta)
}

func (m *NodeGeometry) trimToOrbitRule(delta polar.Polar) polar.Polar {
	rule := m.OrbitRule()
	if rule == nil {
		return delta
	}
	for neighborID := range m.NeighborKinds() {
		if m.IsOutTarget(neighborID) {
			continue
		}
		have, ok := m.DeltaFrom(neighborID)
		if !ok {
			continue
		}
		want := polar.Compose(have, delta)
		delta = polar.Between(have, rule.TrimDelta(have, want))
	}
	return delta
}

func (m *NodeGeometry) trimToEqualOutLengths(delta polar.Polar) polar.Polar {
	if m.SelfKind() != SharedLengthKind {
		return delta
	}

	longest, shortest, count := 0.0, 0.0, 0
	for _, to := range m.OutTargets() {
		d, ok := m.DeltaTo(to)
		if !ok {
			continue
		}
		if count == 0 || d.R > longest {
			longest = d.R
		}
		if count == 0 || d.R < shortest {
			shortest = d.R
		}
		count++
	}
	if count < 2 || longest == shortest {
		return delta
	}

	return polar.Polar{R: delta.R - (longest - shortest), Phi: delta.Phi, Theta: delta.Theta}
}
