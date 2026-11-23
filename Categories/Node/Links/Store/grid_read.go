package Store

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"io/fs"
	"maps"
	"slices"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

func writeBase(b *bytes.Buffer, base map[string]Links.Base) {
	le := binary.LittleEndian
	_ = binary.Write(b, le, uint32(len(base)*len(baseRows)))
	for _, from := range slices.Sorted(maps.Keys(base)) {
		for _, r := range baseRows {
			g := newRow(BandIn+from, r.name, r.isInt, 1)
			g.set(0, r.cell(base[from]))
			putRow(b, g)
		}
	}
}

func readBase(r io.Reader, ver uint32) ([]GridRow, error) {
	var n uint32
	if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
		return nil, err
	}
	out := make([]GridRow, n)
	for i := range out {
		var err error
		if out[i], err = getRow(r, ver, 1); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func ReadGrid(path string) (Links.Grid, bool, error) {
	h, rows, err := ReadGridRows(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Links.Grid{}, false, nil
	}
	if err != nil {
		return Links.Grid{}, false, err
	}
	g := Links.Grid{ID: h.ID, Level: h.Level, Shown: h.Shown, Cols: make([]Links.Col, len(h.Steps))}
	for i, s := range h.Steps {
		g.Cols[i].Step = s
	}
	g.Base = map[string]Links.Base{}
	for _, r := range h.Base {
		from := strings.TrimPrefix(r.Band, BandIn)
		b := g.Base[from]
		for _, br := range baseRows {
			if br.name == r.Name {
				br.put(&b, r.Cell(0))
			}
		}
		g.Base[from] = b
	}
	for from, old := range legacyBase(h.Base) {
		b := g.Base[from]
		b.Origin.S = 100
		b.Origin.Tick = Links.Ticks{Phi: Links.LegacyTicks(old[0], 100), Theta: Links.LegacyTicks(old[1], 100)}
		g.Base[from] = b
	}
	byBand := map[string]map[string]GridRow{}
	var bands []string
	for _, r := range rows {
		if byBand[r.Band] == nil {
			byBand[r.Band] = map[string]GridRow{}
			bands = append(bands, r.Band)
		}
		byBand[r.Band][r.Name] = r
	}
	setRows(byBand[BandNode], nodeRows, func(i int) *Links.Col { return &g.Cols[i] }, len(g.Cols))
	for _, band := range bands {
		vs := byBand[band]
		for i := range g.Cols {
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
			if _, ok := vs[RowWayTheta]; !ok {
				l.Way.Phi = Links.WayNone
			}
			for _, r := range sumRows {
				r.put(&sum, at(vs, r.name, i))
			}
			c := &g.Cols[i]
			switch {
			case strings.HasPrefix(band, BandIn):
				l.From, l.To = strings.TrimPrefix(band, BandIn), h.ID
				c.In = append(c.In, Links.InCell{Link: l, Sum: sum})
			case strings.HasPrefix(band, BandOut):
				l.From, l.To = h.ID, strings.TrimPrefix(band, BandOut)
				o := Links.OutCell{Link: l, Sum: sum}
				for _, r := range outRows {
					r.put(&o, at(vs, r.name, i))
				}
				c.Out = append(c.Out, o)
			}
		}
	}
	for i := range g.Cols {
		sortIn(g.Cols[i].In)
	}
	if h.Ver < gapOriginsVersion {
		gapOrigins(g)
	}
	return g, true, nil
}

func gapOrigins(g Links.Grid) {
	shown := g.ShownCol()
	for from, b := range g.Base {
		in, ok := shown.InFrom(from)
		if !ok || in.Link.Picked {
			continue
		}
		d := in.Link.Delta
		b.Origin = Links.BaseOf(b.Origin, Links.Ints{Phi: b.Total.Phi + d.Phi, Theta: b.Total.Theta + d.Theta, R: b.Total.R + d.R}).Origin
		g.Base[from] = b
	}
}

func legacyBase(rows []GridRow) map[string][2]float64 {
	has, old := map[string]bool{}, map[string][2]float64{}
	for _, r := range rows {
		from := strings.TrimPrefix(r.Band, BandIn)
		o := old[from]
		switch r.Name {
		case RowPhiTick:
			has[from] = true
		case rowPhiOld:
			o[0] = r.V[0]
		case rowThetaOld:
			o[1] = r.V[0]
		}
		old[from] = o
	}
	for from := range has {
		delete(old, from)
	}
	return old
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
