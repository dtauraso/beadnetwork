package owners

import "github.com/dtauraso/wirefold/nodes/Wiring/polarindex"

type Deltas struct {
	baseTo    map[string]polarindex.Index
	dragTo    map[string]polarindex.Index
	constants polarindex.SceneConstants
}

func NewDeltas() Deltas { return Deltas{} }

func (d *Deltas) SetConstants(sc polarindex.SceneConstants) { d.constants = sc }

func (d *Deltas) SetBaseDeltaTo(otherID string, idx polarindex.Index) {
	if d.baseTo == nil {
		d.baseTo = map[string]polarindex.Index{}
	}
	d.baseTo[otherID] = idx
}

func (d *Deltas) SetDragDeltaTo(otherID string, idx polarindex.Index) {
	if d.dragTo == nil {
		d.dragTo = map[string]polarindex.Index{}
	}
	d.dragTo[otherID] = idx
}

func (d *Deltas) DeltaTo(otherID string) (polarindex.Index, bool) {
	base, ok := d.baseTo[otherID]
	if !ok {
		return polarindex.Index{}, false
	}
	return polarindex.Compose(base, d.dragTo[otherID], d.constants), true
}

func (d *Deltas) DragDeltaTo(otherID string) (polarindex.Index, bool) {
	if _, ok := d.baseTo[otherID]; !ok {
		return polarindex.Index{}, false
	}
	return d.dragTo[otherID], true
}

func (d *Deltas) DeltaFrom(otherID string) (polarindex.Index, bool) {
	idx, ok := d.DeltaTo(otherID)
	if !ok {
		return polarindex.Index{}, false
	}
	return polarindex.Neg(idx), true
}

func (d *Deltas) ShiftSelfBy(delta polarindex.Index) {
	if d.dragTo == nil {
		d.dragTo = map[string]polarindex.Index{}
	}
	for id := range d.baseTo {
		d.dragTo[id] = polarindex.Compose(d.dragTo[id], polarindex.Neg(delta), d.constants)
	}
}

func (d *Deltas) ShiftOtherBy(otherID string, delta polarindex.Index) {
	if _, ok := d.baseTo[otherID]; !ok {
		return
	}
	if d.dragTo == nil {
		d.dragTo = map[string]polarindex.Index{}
	}
	d.dragTo[otherID] = polarindex.Compose(d.dragTo[otherID], delta, d.constants)
}

func (d *Deltas) DeltaIDs() []string {
	ids := make([]string, 0, len(d.baseTo))
	for id := range d.baseTo {
		ids = append(ids, id)
	}
	return ids
}
