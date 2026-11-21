package Links

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"io/fs"
	"maps"
	"slices"
	"strings"
)

func writeBase(b *bytes.Buffer, base map[string]Base) {
	le := binary.LittleEndian
	_ = binary.Write(b, le, uint32(len(base)*len(baseRows)))
	for _, from := range slices.Sorted(maps.Keys(base)) {
		for _, r := range baseRows {
			putStr(b, BandIn+from)
			putStr(b, r.name)
			_ = binary.Write(b, le, r.get(base[from]))
		}
	}
}

func readBase(r io.Reader) ([]GridRow, error) {
	var n uint32
	if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
		return nil, err
	}
	out := make([]GridRow, n)
	for i := range out {
		out[i].Band, _ = getStr(r)
		out[i].Name, _ = getStr(r)
		out[i].V = make([]float64, 1)
		if err := binary.Read(r, binary.LittleEndian, out[i].V); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func ReadGrid(path string) (Grid, bool, error) {
	h, rows, err := ReadGridRows(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Grid{}, false, nil
	}
	if err != nil {
		return Grid{}, false, err
	}
	g := Grid{ID: h.ID, Level: h.Level, Shown: h.Shown, Cols: make([]Col, len(h.Steps))}
	for i, s := range h.Steps {
		g.Cols[i].Step = s
	}
	g.Base = map[string]Base{}
	for _, r := range h.Base {
		from := strings.TrimPrefix(r.Band, BandIn)
		b := g.Base[from]
		for _, br := range baseRows {
			if br.name == r.Name {
				br.set(&b, r.V[0])
			}
		}
		g.Base[from] = b
	}
	for from, old := range legacyBase(h.Base) {
		b := g.Base[from]
		b.Origin.S = 100
		b.Origin.Tick = Ticks{LegacyTicks(old[0], 100), LegacyTicks(old[1], 100)}
		g.Base[from] = b
	}
	byBand := map[string]map[string][]float64{}
	var bands []string
	for _, r := range rows {
		if byBand[r.Band] == nil {
			byBand[r.Band] = map[string][]float64{}
			bands = append(bands, r.Band)
		}
		byBand[r.Band][r.Name] = r.V
	}
	setRows(byBand[BandNode], nodeRows, func(i int) *Col { return &g.Cols[i] }, len(g.Cols))
	for _, band := range bands {
		vs := byBand[band]
		for i := range g.Cols {
			if p := vs[rowPresent]; p == nil || p[i] == 0 {
				continue
			}
			var l Link
			var sum Angles
			for _, r := range linkRows {
				r.set(&l, at(vs, r.name, i))
			}
			if vs[RowPhiTick] == nil {
				l.Start.Tick = Ticks{LegacyTicks(at(vs, rowPhiOld, i), l.Start.S), LegacyTicks(at(vs, rowThetaOld, i), l.Start.S)}
			}
			for _, r := range sumRows {
				r.set(&sum, at(vs, r.name, i))
			}
			c := &g.Cols[i]
			switch {
			case strings.HasPrefix(band, BandIn):
				l.From, l.To = strings.TrimPrefix(band, BandIn), h.ID
				c.In = append(c.In, InCell{Link: l, Sum: sum})
			case strings.HasPrefix(band, BandOut):
				l.From, l.To = h.ID, strings.TrimPrefix(band, BandOut)
				o :=OutCell{Link: l, Sum: sum}
				for _, r := range outRows {
					r.set(&o, at(vs, r.name, i))
				}
				c.Out = append(c.Out, o)
			}
		}
	}
	for i := range g.Cols {
		sortIn(g.Cols[i].In)
	}
	if h.Ver < gapOriginsVersion {
		g.gapOrigins()
	}
	return g, true, nil
}

func (g Grid) gapOrigins() {
	shown := g.ShownCol()
	for from, b := range g.Base {
		in, ok := shown.InFrom(from)
		if !ok || in.Link.Picked {
			continue
		}
		d := in.Link.Delta
		b.Origin = baseOf(b.Origin, Ints{Phi: b.Total.Phi + d.Phi, Theta: b.Total.Theta + d.Theta, R: b.Total.R + d.R}).Origin
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

func at(vs map[string][]float64, name string, i int) float64 {
	if v := vs[name]; i < len(v) {
		return v[i]
	}
	return 0
}

func setRows[T any](vs map[string][]float64, rs []row[T], t func(i int) *T, n int) {
	for i := range n {
		for _, r := range rs {
			r.set(t(i), at(vs, r.name, i))
		}
	}
}

func sortIn(in []InCell) {
	ls := make([]Link, len(in))
	for i, c := range in {
		ls[i] = c.Link
	}
	SortLinks(ls)
	sums := map[string]Angles{}
	for _, c := range in {
		sums[c.Link.From] = c.Sum
	}
	for i, l := range ls {
		in[i] = InCell{Link: l, Sum: sums[l.From]}
	}
}
