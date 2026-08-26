package Topology

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
	beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"
	edge "github.com/dtauraso/beadnetwork/Categories/Node/Edge"
)

func NodeTypes(spec TopoSpec) map[string]string {
	nodeType := map[string]string{}
	for _, n := range spec.Nodes {
		nodeType[n.ID] = n.Type
	}
	return nodeType
}

func AllocateBeadLines(spec TopoSpec, nodeGeoms map[string]NodeBuf.NodeGeom) (
	destRun map[string]*beadanimation.BeadLine,
	edgeRun map[string]*beadanimation.BeadLine,
	edgeEndpoints map[string]edge.EdgeEndpoints,
) {
	destRun = map[string]*beadanimation.BeadLine{}
	edgeRun = map[string]*beadanimation.BeadLine{}
	edgeEndpoints = map[string]edge.EdgeEndpoints{}
	for _, e := range spec.Edges {
		if e.SourceHandle == "" && e.TargetHandle == "" {

			edgeEndpoints[e.Label] = edge.EdgeEndpoints{Source: e.Source, Target: e.Target}
			continue
		}

		destKey := e.Target + "." + e.TargetHandle

		if _, exists := destRun[destKey]; exists {
			panic("AllocateBeadLines: two edges target " + destKey + " — validateNoFanIn should have rejected this fan-in at parse")
		}

		pw := beadanimation.NewBeadLine()
		pw.Owner = e.Source
		pw.Edge = e.Label
		pw.Target = e.Target
		pw.TargetHandle = e.TargetHandle
		destRun[destKey] = pw
		edgeRun[e.Label] = pw
		edgeEndpoints[e.Label] = edge.EdgeEndpoints{
			Source: e.Source, Target: e.Target,
			SourceHandle: e.SourceHandle, TargetHandle: e.TargetHandle,
		}
	}
	return destRun, edgeRun, edgeEndpoints
}

func AllocateVectorChannels(spec TopoSpec) (vectorOutByNode, vectorInByNode map[string]chan TiltPanel.TiltVectorMsg) {
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

func BuildEdgeMaps(spec TopoSpec, nodeType map[string]string, kindBroadcastPorts map[string]map[string]bool) (inbound map[string]map[string]string, outbound map[string]map[string][]string, outboundHandle map[string]map[string][]string) {
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

func NodeSendRule(n Node, port string) beadanimation.SendRule {
	if n.Data == nil || n.Data.SendRules == nil {
		return beadanimation.RuleConsumeGated
	}

	rule, _ := beadanimation.ParseSendRule(n.Data.SendRules[port])
	return rule
}
