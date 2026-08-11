// build_edge_maps.go — the id→type map and per-kind Broadcast port set buildNodes wires
// each node's ports from (needed for sourceHandle normalization). The inbound/outbound
// edge maps themselves (buildEdgeMaps) moved to nodes/Wiring/topoderive — this function
// stayed behind because it reads the package-level Registry, whose value type NodeBuilder
// is a Wiring type topoderive cannot import.

package dispatch

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
)

// buildTypeMaps builds the id→type map and per-kind Broadcast port set (needed
// for sourceHandle normalization in buildEdgeMaps). It reads the package-level
// Registry, not just spec, for the per-kind Broadcast port set.
func buildTypeMaps(spec loadspec.TopoSpec) (nodeType map[string]string, kindBroadcastPorts map[string]map[string]bool) {
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
