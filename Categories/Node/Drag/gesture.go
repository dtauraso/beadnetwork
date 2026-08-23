package Drag

import "github.com/dtauraso/wirefold/Categories/Node/nodegeom"

type Gesture struct {
	Node string

	StartCenter nodegeom.Vec3

	GrabOffset nodegeom.Vec3
}

func (d *Gesture) Clear() {
	d.Node = ""
	d.GrabOffset = nodegeom.Vec3{}
}

func (d *Gesture) Holding() bool { return d.Node != "" }

func (d *Gesture) GrabAt(hit nodegeom.Vec3) {
	d.GrabOffset = d.StartCenter.Sub(hit)
}

func (d *Gesture) TargetFor(hit nodegeom.Vec3) nodegeom.Vec3 { return hit.Add(d.GrabOffset) }
