package Topology

import (
	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

func LinkFrame(spec TopoSpec) Links.Frame {
	rho := spec.Rho
	if rho <= 0 && len(spec.Nodes) > 0 {
		rho = NodeBuf.NodeRadius(spec.Nodes[0].Type)
	}
	c := spec.Constants
	return Links.Frame{Rho: rho, Cr: c.ConstantR, Tphi: c.MaxIndexPhi, Ttheta: c.MaxIndexTheta}
}

func LinkSeed(spec TopoSpec, sceneS int) Links.Seed {
	sd := Links.Seed{Frame: LinkFrame(spec), Places: map[string]Links.Place{}, Scales: map[string]int{}, Links: spec.Links}
	for _, n := range spec.Nodes {
		if n.At != nil {
			sd.Places[n.ID] = *n.At
		}
		sd.Scales[n.ID] = sceneS
		if n.S != nil && *n.S >= 1 {
			sd.Scales[n.ID] = *n.S
		}
	}
	return sd
}
