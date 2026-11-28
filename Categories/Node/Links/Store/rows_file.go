package Store

import (
	"fmt"
	"math"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	BandNode   = "node"
	BandIn     = "in "
	BandOut    = "out "
	rowPresent = "present"
)

type GridRow struct {
	Band, Name string
	Int        bool
	V          []float64
	I          []int64
}

type Col struct {
	Step       int
	Ref        Links.Ref
	At         Links.Place
	Pos, Label Links.Vec3
	In         []Links.InCell
	Out        []Links.OutCell
	Counted    bool
}

func (c Col) InFrom(from string) (Links.InCell, bool) {
	for _, in := range c.In {
		if in.Link.From == from {
			return in, true
		}
	}
	return Links.InCell{}, false
}

func colOf(n Links.Node) Col {
	return Col{Ref: n.Ref, At: n.At, Pos: n.Pos, Label: n.Label, In: n.In, Out: n.Out}
}

func (c Col) node(id string) Links.Node {
	return Links.Node{ID: id, Ref: c.Ref, At: c.At, Pos: c.Pos, Label: c.Label, In: c.In, Out: c.Out}
}

func rowsOf[T any](band string, rs []row[T], at func(c *Col) (T, bool), cols []Col) []GridRow {
	var out []GridRow
	for _, r := range rs {
		gr := newRow(band, r.name, r.isInt, len(cols))
		for i := range cols {
			if t, ok := at(&cols[i]); ok {
				gr.set(i, r.cell(t))
			}
		}
		out = append(out, gr)
	}
	return out
}

func bandRows[C any](band string, cols []Col, cell func(c *Col) (C, bool), link func(C) Links.Link, sum func(C) Links.Angles, extra []row[C]) []GridRow {
	present := newRow(band, rowPresent, true, len(cols))
	for i := range cols {
		_, ok := cell(&cols[i])
		present.I[i] = b2i(ok)
	}
	out := []GridRow{present}
	out = append(out, rowsOf(band, linkRows, func(c *Col) (Links.Link, bool) { x, ok := cell(c); return link(x), ok }, cols)...)
	out = append(out, rowsOf(band, sumRows, func(c *Col) (Links.Angles, bool) { x, ok := cell(c); return sum(x), ok }, cols)...)
	return append(out, rowsOf(band, extra, cell, cols)...)
}

func colRows(cols []Col) []GridRow {
	rows := rowsOf(BandNode, nodeRows, func(c *Col) (Col, bool) { return *c, true }, cols)
	var ins, outs []string
	seen := map[string]bool{}
	for _, c := range cols {
		for _, in := range c.In {
			if k := BandIn + in.Link.From; !seen[k] {
				seen[k], ins = true, append(ins, in.Link.From)
			}
		}
		for _, o := range c.Out {
			if k := BandOut + o.Link.To; !seen[k] {
				seen[k], outs = true, append(outs, o.Link.To)
			}
		}
	}
	for _, from := range ins {
		cell := func(c *Col) (Links.InCell, bool) { return c.InFrom(from) }
		rows = append(rows, bandRows(BandIn+from, cols, cell, func(x Links.InCell) Links.Link { return x.Link }, func(x Links.InCell) Links.Angles { return x.Sum }, nil)...)
	}
	for _, to := range outs {
		cell := func(c *Col) (Links.OutCell, bool) {
			for _, o := range c.Out {
				if o.Link.To == to {
					return o, true
				}
			}
			return Links.OutCell{}, false
		}
		rows = append(rows, bandRows(BandOut+to, cols, cell, func(x Links.OutCell) Links.Link { return x.Link }, func(x Links.OutCell) Links.Angles { return x.Sum }, outRows)...)
	}
	return rows
}

func colsOf(id string, rows []GridRow, cols []Col) error {
	byBand := map[string]map[string]GridRow{}
	var bands []string
	for _, r := range rows {
		for _, x := range r.V {
			if math.IsNaN(x) {
				return fmt.Errorf("row %s %s holds NaN — every cell is a number the node wrote", r.Band, r.Name)
			}
		}
		if byBand[r.Band] == nil {
			byBand[r.Band] = map[string]GridRow{}
			bands = append(bands, r.Band)
		}
		byBand[r.Band][r.Name] = r
	}
	setRows(byBand[BandNode], nodeRows, func(i int) *Col { return &cols[i] }, len(cols))
	for _, band := range bands {
		vs := byBand[band]
		for i := range cols {
			if at(vs, rowPresent, i).Zero() {
				continue
			}
			var l Links.Link
			var sum Links.Angles
			for _, r := range linkRows {
				r.put(&l, at(vs, r.name, i))
			}
			if _, ok := vs[RowPhiTick]; !ok {
				l.Start.Tick = Links.Ticks{Phi: Links.LegacyTicks(at(vs, rowPhiOld, i).F, l.Start.S), Theta: Links.LegacyTicks(at(vs, rowThetaOld, i).F, l.Start.S)}
			}
			for _, r := range sumRows {
				r.put(&sum, at(vs, r.name, i))
			}
			c := &cols[i]
			switch {
			case strings.HasPrefix(band, BandIn):
				l.From, l.To = strings.TrimPrefix(band, BandIn), id
				c.In = append(c.In, Links.InCell{Link: l, Sum: sum})
			case strings.HasPrefix(band, BandOut):
				l.From, l.To = id, strings.TrimPrefix(band, BandOut)
				o := Links.OutCell{Link: l, Sum: sum}
				for _, r := range outRows {
					r.put(&o, at(vs, r.name, i))
				}
				c.Out = append(c.Out, o)
			}
		}
	}
	for i := range cols {
		sortIn(cols[i].In)
	}
	return nil
}

func at(vs map[string]GridRow, name string, i int) Cell { return vs[name].Cell(i) }

func setRows[T any](vs map[string]GridRow, rs []row[T], t func(i int) *T, n int) {
	for i := range n {
		for _, r := range rs {
			r.put(t(i), at(vs, r.name, i))
		}
	}
}

func sortIn(in []Links.InCell) {
	ls := make([]Links.Link, len(in))
	for i, c := range in {
		ls[i] = c.Link
	}
	SortLinks(ls)
	sums := map[string]Links.Angles{}
	for _, c := range in {
		sums[c.Link.From] = c.Sum
	}
	for i, l := range ls {
		in[i] = Links.InCell{Link: l, Sum: sums[l.From]}
	}
}
