package Links

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const (
	FileGrid    = "grid.bin"
	gridMagic   = "GRID"
	gridVersion = 2
	BandNode    = "node"
	BandIn      = "in "
	BandOut     = "out "
	rowPresent  = "present"
)

type GridRow struct {
	Band, Name string
	V          []float64
}

func rowsOf[T any](band string, rs []row[T], at func(c *Col) (T, bool), cols []Col) []GridRow {
	var out []GridRow
	for _, r := range rs {
		gr := GridRow{Band: band, Name: r.name, V: make([]float64, len(cols))}
		for i := range cols {
			if t, ok := at(&cols[i]); ok {
				gr.V[i] = r.get(t)
			}
		}
		out = append(out, gr)
	}
	return out
}

func bandRows[C any](band string, cols []Col, cell func(c *Col) (C, bool), link func(C) Link, sum func(C) Start, extra []row[C]) []GridRow {
	present := GridRow{Band: band, Name: rowPresent, V: make([]float64, len(cols))}
	for i := range cols {
		_, ok := cell(&cols[i])
		present.V[i] = b2f(ok)
	}
	out := []GridRow{present}
	out = append(out, rowsOf(band, linkRows, func(c *Col) (Link, bool) { x, ok := cell(c); return link(x), ok }, cols)...)
	out = append(out, rowsOf(band, sumRows, func(c *Col) (Start, bool) { x, ok := cell(c); return sum(x), ok }, cols)...)
	return append(out, rowsOf(band, extra, cell, cols)...)
}

func (g Grid) Rows() []GridRow {
	rows := rowsOf(BandNode, nodeRows, func(c *Col) (Col, bool) { return *c, true }, g.Cols)
	var ins, outs []string
	seen := map[string]bool{}
	for _, c := range g.Cols {
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
		cell := func(c *Col) (InCell, bool) { return c.InFrom(from) }
		rows = append(rows, bandRows(BandIn+from, g.Cols, cell, func(x InCell) Link { return x.Link }, func(x InCell) Start { return x.Sum }, nil)...)
	}
	for _, to := range outs {
		cell := func(c *Col) (OutCell, bool) {
			for _, o := range c.Out {
				if o.Link.To == to {
					return o, true
				}
			}
			return OutCell{}, false
		}
		rows = append(rows, bandRows(BandOut+to, g.Cols, cell, func(x OutCell) Link { return x.Link }, func(x OutCell) Start { return x.Sum }, outRows)...)
	}
	return rows
}

func putStr(b *bytes.Buffer, s string) {
	_ = binary.Write(b, binary.LittleEndian, uint16(len(s)))
	b.WriteString(s)
}

func WriteGrid(path string, g Grid) error {
	var b bytes.Buffer
	b.WriteString(gridMagic)
	le := binary.LittleEndian
	_ = binary.Write(&b, le, uint32(gridVersion))
	putStr(&b, g.ID)
	putStr(&b, filepath.ToSlash(g.Level))
	_ = binary.Write(&b, le, int64(g.Shown))
	_ = binary.Write(&b, le, uint32(len(g.Cols)))
	for _, c := range g.Cols {
		_ = binary.Write(&b, le, int64(c.Step))
	}
	rows := g.Rows()
	_ = binary.Write(&b, le, uint32(len(rows)))
	for _, r := range rows {
		putStr(&b, r.Band)
		putStr(&b, r.Name)
		_ = binary.Write(&b, le, r.V)
	}
	writeBase(&b, g.Base)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if was, err := os.ReadFile(path); err == nil && bytes.Equal(was, b.Bytes()) {
		return nil
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b.Bytes(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

type GridHead struct {
	ID, Level string
	Shown     int
	Steps     []int
	Base      []GridRow
}

func getStr(r io.Reader) (string, error) {
	var n uint16
	if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
		return "", err
	}
	b := make([]byte, n)
	_, err := io.ReadFull(r, b)
	return string(b), err
}

func ReadGridRows(path string) (GridHead, []GridRow, error) {
	var h GridHead
	data, err := os.ReadFile(path)
	if err != nil {
		return h, nil, err
	}
	if !strings.HasPrefix(string(data), gridMagic) {
		return h, nil, fmt.Errorf("Links.ReadGridRows: %s does not start with %q, so it is not a node grid", path, gridMagic)
	}
	r, le := bytes.NewReader(data[len(gridMagic):]), binary.LittleEndian
	var ver, ncols, nrows uint32
	var shown int64
	_ = binary.Read(r, le, &ver)
	if ver != 1 && ver != gridVersion {
		return h, nil, fmt.Errorf("Links.ReadGridRows: %s is grid version %d, and this build reads versions 1 and %d", path, ver, gridVersion)
	}
	h.ID, _ = getStr(r)
	h.Level, _ = getStr(r)
	_ = binary.Read(r, le, &shown)
	h.Shown = int(shown)
	_ = binary.Read(r, le, &ncols)
	steps := make([]int64, ncols)
	_ = binary.Read(r, le, steps)
	for _, s := range steps {
		h.Steps = append(h.Steps, int(s))
	}
	if err := binary.Read(r, le, &nrows); err != nil {
		return h, nil, fmt.Errorf("Links.ReadGridRows: %s ends inside its header: %w", path, err)
	}
	rows := make([]GridRow, nrows)
	for i := range rows {
		rows[i].Band, _ = getStr(r)
		rows[i].Name, _ = getStr(r)
		rows[i].V = make([]float64, ncols)
		if err := binary.Read(r, le, rows[i].V); err != nil {
			return h, nil, fmt.Errorf("Links.ReadGridRows: %s ends inside row %d (%s %s): %w", path, i, rows[i].Band, rows[i].Name, err)
		}
	}
	if ver >= 2 {
		if h.Base, err = readBase(r); err != nil {
			return h, nil, fmt.Errorf("Links.ReadGridRows: %s ends inside its running totals: %w", path, err)
		}
	}
	for _, v := range append(rows, h.Base...) {
		for _, x := range v.V {
			if math.IsNaN(x) {
				return h, nil, fmt.Errorf("Links.ReadGridRows: %s row %s %s holds NaN — every cell is a number the node wrote", path, v.Band, v.Name)
			}
		}
	}
	return h, rows, nil
}
