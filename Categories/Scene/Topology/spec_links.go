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

func LinkSeed(spec TopoSpec, sceneS, tickMs int) Links.Seed {
	sd := Links.Seed{Frame: LinkFrame(spec), Places: map[string]Links.Place{}, S: sceneS, Refs: map[string]Links.Ref{}, Links: spec.Links, TickMs: tickMs}
	for _, n := range spec.Nodes {
		sd.Refs[n.ID] = Links.Ref{Phi: n.RefPhi, Theta: n.RefTheta, S: n.RefS}
		Links.MustBeRefCount(n.ID, sd.Refs[n.ID])
		if n.At != nil {
			sd.Places[n.ID] = *n.At
		}
	}
	return sd
}
