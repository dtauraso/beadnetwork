package scenebuild

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/TiltPanel"
	NodeBuf "github.com/dtauraso/wirefold/Categories/Node"
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	edge "github.com/dtauraso/wirefold/Categories/Node/Edge"
	"github.com/dtauraso/wirefold/Categories/Scene/loadspec"
)

func AllocateBeadLines(spec loadspec.TopoSpec, nodeGeoms map[string]NodeBuf.NodeGeom) (
	destRun map[string]*beadanimation.BeadLine,
	edgeRun map[string]*beadanimation.BeadLine,
	edgeEndpoints map[string]edge.EdgeEndpoints,
) {
	destRun = map[string]*beadanimation.BeadLine{}
	edgeRun = map[string]*beadanimation.BeadLine{}
	edgeEndpoints = map[string]edge.EdgeEndpoints{}
	for _, e := range spec.Edges {
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
