package owners

import "github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"

type Deltas struct {
	toOther map[string]polar.Polar
}

func NewDeltas() Deltas { return Deltas{} }

func (d *Deltas) SetDeltaTo(otherID string, p polar.Polar) {
	if d.toOther == nil {
		d.toOther = map[string]polar.Polar{}
	}
	d.toOther[otherID] = p
}

func (d *Deltas) DeltaTo(otherID string) (polar.Polar, bool) {
	p, ok := d.toOther[otherID]
	return p, ok
}

func (d *Deltas) DeltaFrom(otherID string) (polar.Polar, bool) {
	p, ok := d.toOther[otherID]
	if !ok {
		return polar.Polar{}, false
	}
	return p.Neg(), true
}

func (d *Deltas) ShiftSelfBy(delta polar.Polar) {
	for id, p := range d.toOther {
		d.toOther[id] = polar.Compose(p, delta.Neg())
	}
}

func (d *Deltas) ShiftOtherBy(otherID string, delta polar.Polar) {
	p, ok := d.toOther[otherID]
	if !ok {
		return
	}
	d.toOther[otherID] = polar.Compose(p, delta)
}

func (d *Deltas) All() []polar.Polar {
	out := make([]polar.Polar, 0, len(d.toOther))
	for _, p := range d.toOther {
		out = append(out, p)
	}
	return out
}

func (d *Deltas) DeltaIDs() []string {
	ids := make([]string, 0, len(d.toOther))
	for id := range d.toOther {
		ids = append(ids, id)
	}
	return ids
}
