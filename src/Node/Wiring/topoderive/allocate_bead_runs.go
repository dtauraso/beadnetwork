package topoderive

import (
	"github.com/dtauraso/wirefold/src/Input/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/loadspec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodegeom"
	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
)

func AllocateBeadLines(spec loadspec.TopoSpec, nodeGeoms map[string]nodegeom.NodeGeom) (
	destRun map[string]*beadanimation.BeadLine,
	edgeRun loadspec.BeadLineRegistry,
	edgeEndpoints map[string]inputcodec.EdgeEndpoints,
) {
	destRun = map[string]*beadanimation.BeadLine{}
	edgeRun = loadspec.BeadLineRegistry{}
	edgeEndpoints = map[string]inputcodec.EdgeEndpoints{}
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
		edgeEndpoints[e.Label] = inputcodec.EdgeEndpoints{
			Source: e.Source, Target: e.Target,
			SourceHandle: e.SourceHandle, TargetHandle: e.TargetHandle,
		}
	}
	return destRun, edgeRun, edgeEndpoints
}
