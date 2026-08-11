// edge_maps.go — the inbound/outbound edge-direction maps buildNodes wires each node's
// ports from, lifted out of nodes/Wiring/build_edge_maps.go. buildTypeMaps stayed behind
// in nodes/Wiring (it reads the package-level Registry, whose value type NodeBuilder is a
// Wiring type) — its two outputs are still passed in here as plain map types.
package topoderive

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/loadspec"
)

// BuildEdgeMaps builds the inbound and outbound edge maps.
//   - inbound:  target node id → port name → destKey ("destNode.destPort")
//   - outbound: source node id → port name → []edge label
//   - outboundHandle: source node id → port name → []sourceHandle (indexed, same order as outbound)
//
// For Broadcast ports, sourceHandle may be "<portName><index>" — normalize to portName.
func BuildEdgeMaps(spec loadspec.TopoSpec, nodeType map[string]string, kindBroadcastPorts map[string]map[string]bool) (inbound map[string]map[string]string, outbound map[string]map[string][]string, outboundHandle map[string]map[string][]string) {
	inbound = map[string]map[string]string{}
	outbound = map[string]map[string][]string{}
	outboundHandle = map[string]map[string][]string{}
	for _, e := range spec.Edges {
		if inbound[e.Target] == nil {
			inbound[e.Target] = map[string]string{}
		}
		if outbound[e.Source] == nil {
			outbound[e.Source] = map[string][]string{}
		}
		if outboundHandle[e.Source] == nil {
			outboundHandle[e.Source] = map[string][]string{}
		}
		inbound[e.Target][e.TargetHandle] = e.Target + "." + e.TargetHandle
		srcKey := e.SourceHandle
		if base, isMulti := loadspec.BroadcastBaseName(e.SourceHandle, nodeType[e.Source], kindBroadcastPorts); isMulti {
			srcKey = base
		}
		outbound[e.Source][srcKey] = append(outbound[e.Source][srcKey], e.Label)
		outboundHandle[e.Source][srcKey] = append(outboundHandle[e.Source][srcKey], e.SourceHandle)
	}
	return inbound, outbound, outboundHandle
}
