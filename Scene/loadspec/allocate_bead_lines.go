package loadspec

import (
	beadanimation "github.com/dtauraso/wirefold/Node/BeadAnimation"
	edge "github.com/dtauraso/wirefold/Node/Edge"
	"github.com/dtauraso/wirefold/Node/nodegeom"
)

func (spec TopoSpec) AllocateBeadLines(nodeGeoms map[string]nodegeom.NodeGeom) (
	destRun map[string]*beadanimation.BeadLine,
	edgeRun BeadLineRegistry,
	edgeEndpoints map[string]edge.EdgeEndpoints,
) {
	destRun = map[string]*beadanimation.BeadLine{}
	edgeRun = BeadLineRegistry{}
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
