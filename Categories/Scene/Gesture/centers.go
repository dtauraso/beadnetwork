package Gesture

import (
	FitButton "github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/FitButton"
	Node "github.com/dtauraso/beadnetwork/Categories/Node"
	Camera "github.com/dtauraso/beadnetwork/Categories/Scene/Camera"
	Drag "github.com/dtauraso/beadnetwork/Categories/Scene/Drag"
)

func centerOfForMove(f func(id string) (Vec3, bool)) func(id string) (Node.Vec3, bool) {
	return func(id string) (Node.Vec3, bool) {
		c, ok := f(id)
		return Node.Vec3(c), ok
	}
}

func heldCenters(d Deps) map[string]Node.Vec3 {
	return Node.HeldCenters(d.MR.NodeGeoms(), centerOfForMove(d.MR.CenterOf))
}

func centersForFit(in map[string]Node.Vec3) map[string]FitButton.Vec3 {
	out := make(map[string]FitButton.Vec3, len(in))
	for id, c := range in {
		out[id] = FitButton.Vec3(c)
	}
	return out
}

func centersForCamera(in map[string]Node.Vec3) map[string]Camera.Vec3 {
	out := make(map[string]Camera.Vec3, len(in))
	for id, c := range in {
		out[id] = Camera.Vec3(c)
	}
	return out
}

func centersForDrag(in map[string]Node.Vec3) map[string]Drag.Vec3 {
	out := make(map[string]Drag.Vec3, len(in))
	for id, c := range in {
		out[id] = Drag.Vec3(c)
	}
	return out
}
