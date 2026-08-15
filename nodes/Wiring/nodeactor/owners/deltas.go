package owners

import "github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"

type Deltas struct {
	baseTo map[string]polar.Polar
	dragTo map[string]polar.Polar
}

func NewDeltas() Deltas { return Deltas{} }

func (d *Deltas) SetBaseDeltaTo(otherID string, p polar.Polar) {
	if d.baseTo == nil {
		d.baseTo = map[string]polar.Polar{}
	}
	d.baseTo[otherID] = p
}

func (d *Deltas) SetDragDeltaTo(otherID string, p polar.Polar) {
	if d.dragTo == nil {
		d.dragTo = map[string]polar.Polar{}
	}
	d.dragTo[otherID] = p
}

func (d *Deltas) DeltaTo(otherID string) (polar.Polar, bool) {
	base, ok := d.baseTo[otherID]
	if !ok {
		return polar.Polar{}, false
	}
	return polar.Compose(base, d.dragTo[otherID]), true
}

func (d *Deltas) DragDeltaTo(otherID string) (polar.Polar, bool) {
	if _, ok := d.baseTo[otherID]; !ok {
		return polar.Polar{}, false
	}
	return d.dragTo[otherID], true
}

func (d *Deltas) DeltaFrom(otherID string) (polar.Polar, bool) {
	p, ok := d.DeltaTo(otherID)
	if !ok {
		return polar.Polar{}, false
	}
	return p.Neg(), true
}

func (d *Deltas) ShiftSelfBy(delta polar.Polar) {
	if d.dragTo == nil {
		d.dragTo = map[string]polar.Polar{}
	}
	for id := range d.baseTo {
		d.dragTo[id] = polar.Compose(d.dragTo[id], delta.Neg())
	}
}

func (d *Deltas) ShiftOtherBy(otherID string, delta polar.Polar) {
	if _, ok := d.baseTo[otherID]; !ok {
		return
	}
	if d.dragTo == nil {
		d.dragTo = map[string]polar.Polar{}
	}
	d.dragTo[otherID] = polar.Compose(d.dragTo[otherID], delta)
}

func (d *Deltas) DeltaIDs() []string {
	ids := make([]string, 0, len(d.baseTo))
	for id := range d.baseTo {
		ids = append(ids, id)
	}
	return ids
}
