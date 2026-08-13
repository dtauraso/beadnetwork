package owners

import "github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"

// Deltas is a node's own side of every edge it touches: the triple FROM this
// node TO the node at the other end, one per incident edge, in the same polar
// pole convention as the node's own point.
//
// A + D = B — this node's point plus the triple IS the other end's point, added
// component by component — so a node reaches its neighbour without reading
// anything the neighbour owns. Every operation here is arithmetic on r, phi and
// theta; nothing converts, so no cartesian copy of D exists to drift from it.
//
// The triple is stored FROM SELF for an in-edge as well as an out-edge, which
// is what makes a move uniform: when this node moves by Δ and the other end does
// not, every incident triple loses the same Δ, whichever way the edge points.
// The edge's own D, source to target, is the out entry as-is and the negation
// of the in entry.
//
// Only the out entries are persisted, under this node's own edge files. An in
// entry is what its SOURCE told this node, and is re-sent whenever that source
// moves.
type Deltas struct {
	toOther map[string]polar.Polar
}

func NewDeltas() Deltas { return Deltas{} }

// SetDeltaTo records the triple from this node to other.
func (d *Deltas) SetDeltaTo(otherID string, p polar.Polar) {
	if d.toOther == nil {
		d.toOther = map[string]polar.Polar{}
	}
	d.toOther[otherID] = p
}

// DeltaTo is the triple from this node to other.
func (d *Deltas) DeltaTo(otherID string) (polar.Polar, bool) {
	p, ok := d.toOther[otherID]
	return p, ok
}

// DeltaFrom is the vector from other to this node — the same side of the same
// triangle, pointing the other way. It is what an edge INTO this node holds.
func (d *Deltas) DeltaFrom(otherID string) (polar.Polar, bool) {
	p, ok := d.toOther[otherID]
	if !ok {
		return polar.Polar{}, false
	}
	return p.Neg(), true
}

// ShiftSelfBy is what a move of THIS node does to every side it touches: the
// other end did not move, so each vector loses the whole of Δ.
func (d *Deltas) ShiftSelfBy(delta polar.Polar) {
	for id, p := range d.toOther {
		d.toOther[id] = polar.Compose(p, delta.Neg())
	}
}

// ShiftOtherBy is what a move of the node at the OTHER end does to the side
// between them: this node did not move, so the vector to it gains the whole of
// Δ. It runs on what that node told this one, never on a value read out from
// under it.
func (d *Deltas) ShiftOtherBy(otherID string, delta polar.Polar) {
	p, ok := d.toOther[otherID]
	if !ok {
		return
	}
	d.toOther[otherID] = polar.Compose(p, delta)
}

// DeltaIDs is every node at the other end of an incident edge.
func (d *Deltas) DeltaIDs() []string {
	ids := make([]string, 0, len(d.toOther))
	for id := range d.toOther {
		ids = append(ids, id)
	}
	return ids
}
