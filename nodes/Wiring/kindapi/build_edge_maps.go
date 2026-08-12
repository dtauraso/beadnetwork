package kindapi

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
)

func BuildTypeMaps(spec loadspec.TopoSpec) (nodeType map[string]string, kindBroadcastPorts map[string]map[string]bool) {
	nodeType = map[string]string{}
	for _, n := range spec.Nodes {
		nodeType[n.ID] = n.Type
	}
	kindBroadcastPorts = map[string]map[string]bool{}
	for kind, bind := range Registry {
		outMultis := map[string]bool{}
		for _, p := range bind.Ports {
			if p.Dir == portwiring.PortBroadcast {
				outMultis[p.Name] = true
			}
		}
		kindBroadcastPorts[kind] = outMultis
	}
	return nodeType, kindBroadcastPorts
}
