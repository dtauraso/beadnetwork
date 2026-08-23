package Node

type DragGesture struct {
	Node string

	StartCenter Vec3

	GrabOffset Vec3
}

func (d *DragGesture) Clear() {
	d.Node = ""
	d.GrabOffset = Vec3{}
}

func (d *DragGesture) Holding() bool { return d.Node != "" }

func (d *DragGesture) GrabAt(hit Vec3) {
	d.GrabOffset = d.StartCenter.Sub(hit)
}

func (d *DragGesture) TargetFor(hit Vec3) Vec3 { return hit.Add(d.GrabOffset) }
