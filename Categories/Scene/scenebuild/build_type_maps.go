package scenebuild

import (
	"github.com/dtauraso/wirefold/Categories/Scene/loadspec"
)

func BuildTypeMaps(spec loadspec.TopoSpec) (nodeType map[string]string, kindBroadcastPorts map[string]map[string]bool) {
	nodeType = map[string]string{}
	for _, n := range spec.Nodes {
		nodeType[n.ID] = n.Type
	}
	kindBroadcastPorts = map[string]map[string]bool{}
	for kind, ports := range KindPorts {
		outMultis := map[string]bool{}
		for _, p := range ports {
			if p.Dir == PortBroadcast {
				outMultis[p.Name] = true
			}
		}
		kindBroadcastPorts[kind] = outMultis
	}
	return nodeType, kindBroadcastPorts
}
