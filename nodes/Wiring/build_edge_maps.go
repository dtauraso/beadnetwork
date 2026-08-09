// build_edge_maps.go — the id/type and edge-direction maps buildNodes wires each
// node's ports from: the id→type map and per-kind Broadcast port set (needed for
// sourceHandle normalization), plus the inbound/outbound edge maps themselves.

package Wiring

// buildTypeMaps builds the id→type map and per-kind Broadcast port set (needed
// for sourceHandle normalization in buildEdgeMaps).
func (b *buildCtx) buildTypeMaps() {
	nodeType := map[string]string{}
	for _, n := range b.spec.Nodes {
		nodeType[n.ID] = n.Type
	}
	kindBroadcastPorts := map[string]map[string]bool{}
	for kind, bind := range Registry {
		outMultis := map[string]bool{}
		for _, p := range bind.Ports {
			if p.Dir == PortBroadcast {
				outMultis[p.Name] = true
			}
		}
		kindBroadcastPorts[kind] = outMultis
	}
	b.nodeType = nodeType
	b.kindBroadcastPorts = kindBroadcastPorts
}

// buildEdgeMaps builds the inbound and outbound edge maps.
//   - inbound:  target node id → port name → destKey ("destNode.destPort")
//   - outbound: source node id → port name → []edge label
//   - outboundHandle: source node id → port name → []sourceHandle (indexed, same order as outbound)
//
// For Broadcast ports, sourceHandle may be "<portName><index>" — normalize to portName.
func (b *buildCtx) buildEdgeMaps() {
	inbound := map[string]map[string]string{}
	outbound := map[string]map[string][]string{}
	outboundHandle := map[string]map[string][]string{}
	for _, e := range b.spec.Edges {
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
		if base, isMulti := broadcastBaseName(e.SourceHandle, b.nodeType[e.Source], b.kindBroadcastPorts); isMulti {
			srcKey = base
		}
		outbound[e.Source][srcKey] = append(outbound[e.Source][srcKey], e.Label)
		outboundHandle[e.Source][srcKey] = append(outboundHandle[e.Source][srcKey], e.SourceHandle)
	}
	b.inbound = inbound
	b.outbound = outbound
	b.outboundHandle = outboundHandle
}
