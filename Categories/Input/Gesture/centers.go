package Gesture

import (
	FitButton "github.com/dtauraso/wirefold/Categories/Chrome/Pills/FitButton"
	Drag "github.com/dtauraso/wirefold/Categories/Scene/Drag"
	Node "github.com/dtauraso/wirefold/Categories/Node"
	"github.com/dtauraso/wirefold/Categories/Node/nodegeom"
	Camera "github.com/dtauraso/wirefold/Categories/Scene/Camera"
)

func centerOfForMove(f func(id string) (Vec3, bool)) func(id string) (nodegeom.Vec3, bool) {
	return func(id string) (nodegeom.Vec3, bool) {
		c, ok := f(id)
		return nodegeom.Vec3(c), ok
	}
}

func heldCenters(d Deps) map[string]nodegeom.Vec3 {
	return Node.HeldCenters(d.MR.NodeGeoms(), centerOfForMove(d.MR.CenterOf))
}

func centersForFit(in map[string]nodegeom.Vec3) map[string]FitButton.Vec3 {
	out := make(map[string]FitButton.Vec3, len(in))
	for id, c := range in {
		out[id] = FitButton.Vec3(c)
	}
	return out
}

func centersForCamera(in map[string]nodegeom.Vec3) map[string]Camera.Vec3 {
	out := make(map[string]Camera.Vec3, len(in))
	for id, c := range in {
		out[id] = Camera.Vec3(c)
	}
	return out
}

func centersForDrag(in map[string]nodegeom.Vec3) map[string]Drag.Vec3 {
	out := make(map[string]Drag.Vec3, len(in))
	for id, c := range in {
		out[id] = Drag.Vec3(c)
	}
	return out
}
