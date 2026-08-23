package scenebuild

import (
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/Categories/Scene/loadspec"
)

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
		if base, isMulti := BroadcastBaseName(e.SourceHandle, nodeType[e.Source], kindBroadcastPorts); isMulti {
			srcKey = base
		}
		outbound[e.Source][srcKey] = append(outbound[e.Source][srcKey], e.Label)
		outboundHandle[e.Source][srcKey] = append(outboundHandle[e.Source][srcKey], e.SourceHandle)
	}
	return inbound, outbound, outboundHandle
}

func BroadcastBaseName(handle, kind string, kindBroadcastPorts map[string]map[string]bool) (string, bool) {
	if len(handle) == 0 {
		return handle, false
	}
	last := handle[len(handle)-1]
	if last < '0' || last > '9' {
		return handle, false
	}
	base := handle[:len(handle)-1]
	if kindBroadcastPorts[kind][base] {
		return base, true
	}
	return handle, false
}

func NodeSendRule(n loadspec.Node, port string) beadanimation.SendRule {
	if n.Data == nil || n.Data.SendRules == nil {
		return beadanimation.RuleConsumeGated
	}

	rule, _ := beadanimation.ParseSendRule(n.Data.SendRules[port])
	return rule
}
