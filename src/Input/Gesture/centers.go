package Gesture

import (
	Camera "github.com/dtauraso/wirefold/src/Camera"
	FitButton "github.com/dtauraso/wirefold/src/Chrome/Pills/FitButton"
	Drag "github.com/dtauraso/wirefold/src/Input/Drag"
	moverreg "github.com/dtauraso/wirefold/src/Node/moverreg"
	nodemove "github.com/dtauraso/wirefold/src/Node/nodemove"
)

func centerOfForMove(f func(id string) (moverreg.Vec3, bool)) func(id string) (nodemove.Vec3, bool) {
	return func(id string) (nodemove.Vec3, bool) {
		c, ok := f(id)
		return nodemove.Vec3(c), ok
	}
}

func heldCenters(d Deps) map[string]nodemove.Vec3 {
	return nodemove.HeldCenters(d.MR.NodeGeoms(), centerOfForMove(d.MR.CenterOfNode))
}

func centersForFit(in map[string]nodemove.Vec3) map[string]FitButton.Vec3 {
	out := make(map[string]FitButton.Vec3, len(in))
	for id, c := range in {
		out[id] = FitButton.Vec3(c)
	}
	return out
}

func centersForCamera(in map[string]nodemove.Vec3) map[string]Camera.Vec3 {
	out := make(map[string]Camera.Vec3, len(in))
	for id, c := range in {
		out[id] = Camera.Vec3(c)
	}
	return out
}

func centersForDrag(in map[string]nodemove.Vec3) map[string]Drag.Vec3 {
	out := make(map[string]Drag.Vec3, len(in))
	for id, c := range in {
		out[id] = Drag.Vec3(c)
	}
	return out
}
