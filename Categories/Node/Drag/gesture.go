package Drag

type Gesture struct {
	Node string

	StartCenter Vec3

	GrabOffset Vec3
}

func (d *Gesture) Clear() {
	d.Node = ""
	d.GrabOffset = Vec3{}
}

func (d *Gesture) Holding() bool { return d.Node != "" }

func (d *Gesture) GrabAt(hit Vec3) {
	d.GrabOffset = d.StartCenter.Sub(hit)
}

func (d *Gesture) TargetFor(hit Vec3) Vec3 { return hit.Add(d.GrabOffset) }
