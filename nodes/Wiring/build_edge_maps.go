// build_edge_maps.go — the id/type and edge-direction maps buildNodes wires each
// node's ports from: the id→type map and per-kind Broadcast port set (needed for
// sourceHandle normalization), plus the inbound/outbound edge maps themselves.

package Wiring

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

// buildEdgeMaps builds the inbound and outbound edge maps.
//   - inbound:  target node id → port name → destKey ("destNode.destPort")
//   - outbound: source node id → port name → []edge label
//   - outboundHandle: source node id → port name → []sourceHandle (indexed, same order as outbound)
//
// For Broadcast ports, sourceHandle may be "<portName><index>" — normalize to portName.
func buildEdgeMaps(spec loadspec.TopoSpec, nodeType map[string]string, kindBroadcastPorts map[string]map[string]bool) (inbound map[string]map[string]string, outbound map[string]map[string][]string, outboundHandle map[string]map[string][]string) {
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
