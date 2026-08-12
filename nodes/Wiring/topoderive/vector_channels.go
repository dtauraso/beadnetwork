package topoderive

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
)

func AllocateVectorChannels(spec loadspec.TopoSpec) (vectorOutByNode, vectorInByNode map[string]chan tiltvector.TiltVectorMsg) {
	kindByID := make(map[string]string, len(spec.Nodes))
	for _, n := range spec.Nodes {
		kindByID[n.ID] = n.Type
	}
	vectorOutByNode = map[string]chan tiltvector.TiltVectorMsg{}
	vectorInByNode = map[string]chan tiltvector.TiltVectorMsg{}
	for _, e := range spec.Edges {
		if !tiltvector.KindWantsVectorChannel(kindByID[e.Source]) || !tiltvector.KindWantsVectorChannel(kindByID[e.Target]) {
			continue
		}
		sourceToTargetVectorCh := make(chan tiltvector.TiltVectorMsg, 1)
		vectorOutByNode[e.Source] = sourceToTargetVectorCh
		vectorInByNode[e.Target] = sourceToTargetVectorCh
	}
	return vectorOutByNode, vectorInByNode
}
