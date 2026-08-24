package Node

import "github.com/dtauraso/beadnetwork/Categories/Polar/polarindex"

type Deltas struct {
	baseTo    map[string]polarindex.Offset
	dragTo    map[string]polarindex.Offset
	constants polarindex.SceneConstants
}

func NewDeltas() Deltas { return Deltas{} }

func (d *Deltas) SetConstants(sc polarindex.SceneConstants) { d.constants = sc }

func (d *Deltas) SetBaseDeltaTo(otherID string, off polarindex.Offset) {
	if d.baseTo == nil {
		d.baseTo = map[string]polarindex.Offset{}
	}
	d.baseTo[otherID] = off
}

func (d *Deltas) SetDragDeltaTo(otherID string, off polarindex.Offset) {
	if d.dragTo == nil {
		d.dragTo = map[string]polarindex.Offset{}
	}
	d.dragTo[otherID] = off
}

func (d *Deltas) DeltaTo(otherID string) (polarindex.Offset, bool) {
	base, ok := d.baseTo[otherID]
	if !ok {
		return polarindex.Offset{}, false
	}
	return polarindex.Sum(base, d.dragTo[otherID]), true
}

func (d *Deltas) DragDeltaTo(otherID string) (polarindex.Offset, bool) {
	if _, ok := d.baseTo[otherID]; !ok {
		return polarindex.Offset{}, false
	}
	return d.dragTo[otherID], true
}

func (d *Deltas) DeltaFrom(otherID string) (polarindex.Offset, bool) {
	off, ok := d.DeltaTo(otherID)
	if !ok {
		return polarindex.Offset{}, false
	}
	return polarindex.Neg(off), true
}

func (d *Deltas) ShiftSelfBy(delta polarindex.Offset) {
	if d.dragTo == nil {
		d.dragTo = map[string]polarindex.Offset{}
	}
	for id := range d.baseTo {
		d.dragTo[id] = polarindex.Sum(d.dragTo[id], polarindex.Neg(delta))
	}
}

func (d *Deltas) ShiftOtherBy(otherID string, delta polarindex.Offset) {
	if _, ok := d.baseTo[otherID]; !ok {
		return
	}
	if d.dragTo == nil {
		d.dragTo = map[string]polarindex.Offset{}
	}
	d.dragTo[otherID] = polarindex.Sum(d.dragTo[otherID], delta)
}

func (d *Deltas) DeltaIDs() []string {
	ids := make([]string, 0, len(d.baseTo))
	for id := range d.baseTo {
		ids = append(ids, id)
	}
	return ids
}
