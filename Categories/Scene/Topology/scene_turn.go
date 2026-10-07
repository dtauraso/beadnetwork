package Topology

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"
	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Node/Edge/edgefile"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
)

type turn struct{ phi, theta int }

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func lcm(a, b int) int { return a / gcd(a, b) * b }

func (t turn) with(o turn) turn { return turn{lcm(t.phi, o.phi), lcm(t.theta, o.theta)} }

func (t turn) per(saved turn) turn { return turn{t.phi / saved.phi, t.theta / saved.theta} }

func scaleBy(v *int, f int) {
	if v != nil {
		*v *= f
	}
}

func cardTurn(root string, spec TopoSpec) (turn, bool) {
	for _, n := range spec.Nodes {
		if n.Type == CardKind {
			ticks := CardPanel.TurnTicks * CardPanel.LoadCardScalar(Scenes.CardSFilePath(root), CardPanel.DefaultS)
			return turn{ticks, ticks}, true
		}
	}
	return turn{}, false
}

func applySceneTurn(root string, spec *TopoSpec) {
	saved := turn{spec.Constants.MaxIndexPhi, spec.Constants.MaxIndexTheta}
	spec.savedTurn = saved
	t := saved
	if ticks, ok := cardTurn(root, *spec); ok {
		t = t.with(ticks)
	}
	for _, n := range spec.Nodes {
		if phi, theta, ok := NodeBuf.ReadDragTurn(root, n.ID); ok {
			t = t.with(turn{phi, theta})
		}
	}
	for _, e := range spec.Edges {
		if phi, theta, ok := edgefile.ReadEdgeDragTurn(root, e.Source, e.Label); ok {
			t = t.with(turn{phi, theta})
		}
	}
	spec.Constants.MaxIndexPhi, spec.Constants.MaxIndexTheta = t.phi, t.theta

	f := t.per(saved)
	for i := range spec.Nodes {
		scaleBy(spec.Nodes[i].IndexPhi, f.phi)
		scaleBy(spec.Nodes[i].IndexTheta, f.theta)
	}
	for i := range spec.Edges {
		scaleBy(spec.Edges[i].DeltaIndexPhi, f.phi)
		scaleBy(spec.Edges[i].DeltaIndexTheta, f.theta)
	}
}

func (spec TopoSpec) dragScale(phi, theta int, ok bool) turn {
	from := spec.savedTurn
	if ok {
		from = turn{phi, theta}
	}
	return turn{spec.Constants.MaxIndexPhi, spec.Constants.MaxIndexTheta}.per(from)
}
