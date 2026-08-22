package loadspec

import ()

func (spec TopoSpec) BuildEdgeMaps(nodeType map[string]string, kindBroadcastPorts map[string]map[string]bool) (inbound map[string]map[string]string, outbound map[string]map[string][]string, outboundHandle map[string]map[string][]string) {
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
		if base, isMulti := BroadcastBaseName(e.SourceHandle, nodeType[e.Source], kindBroadcastPorts); isMulti {
			srcKey = base
		}
		outbound[e.Source][srcKey] = append(outbound[e.Source][srcKey], e.Label)
		outboundHandle[e.Source][srcKey] = append(outboundHandle[e.Source][srcKey], e.SourceHandle)
	}
	return inbound, outbound, outboundHandle
}
