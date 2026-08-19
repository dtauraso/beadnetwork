package topoderive

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/tools/topology-vscode/src/TiltPanel"
)

func AllocateVectorChannels(spec loadspec.TopoSpec) (vectorOutByNode, vectorInByNode map[string]chan TiltPanel.TiltVectorMsg) {
	kindByID := make(map[string]string, len(spec.Nodes))
	for _, n := range spec.Nodes {
		kindByID[n.ID] = n.Type
	}
	vectorOutByNode = map[string]chan TiltPanel.TiltVectorMsg{}
	vectorInByNode = map[string]chan TiltPanel.TiltVectorMsg{}
	for _, e := range spec.Edges {
		if !TiltPanel.KindWantsVectorChannel(kindByID[e.Source]) || !TiltPanel.KindWantsVectorChannel(kindByID[e.Target]) {
			continue
		}
		sourceToTargetVectorCh := make(chan TiltPanel.TiltVectorMsg, 1)
		vectorOutByNode[e.Source] = sourceToTargetVectorCh
		vectorInByNode[e.Target] = sourceToTargetVectorCh
	}
	return vectorOutByNode, vectorInByNode
}
