package Store

import (
	"bytes"
	"encoding/binary"
	"io"
)

const typedVersion = 5

type Cell struct {
	Int bool
	I   int64
	F   float64
}

func (c Cell) Zero() bool { return c.I == 0 && c.F == 0 }

func newRow(band, name string, isInt bool, n int) GridRow {
	g := GridRow{Band: band, Name: name, Int: isInt}
	if isInt {
		g.I = make([]int64, n)
	} else {
		g.V = make([]float64, n)
	}
	return g
}

func (g GridRow) Cell(i int) Cell {
	if g.Int {
		if i < len(g.I) {
			return Cell{Int: true, I: g.I[i]}
		}
		return Cell{Int: true}
	}
	if i < len(g.V) {
		return Cell{F: g.V[i]}
	}
	return Cell{}
}

func (g *GridRow) set(i int, c Cell) {
	if g.Int {
		g.I[i] = c.I
	} else {
		g.V[i] = c.F
	}
}

func (r row[T]) cell(t T) Cell {
	if r.isInt {
		return Cell{Int: true, I: r.getI(t)}
	}
	return Cell{F: r.getF(t)}
}

func (r row[T]) put(t *T, c Cell) {
	if r.isInt {
		v := c.I
		if !c.Int {
			v = int64(c.F)
		}
		r.setI(t, v)
		return
	}
	f := c.F
	if c.Int {
		f = float64(c.I)
	}
	r.setF(t, f)
}

func putRow(b *bytes.Buffer, g GridRow) {
	putStr(b, g.Band)
	putStr(b, g.Name)
	if g.Int {
		b.WriteByte(1)
		_ = binary.Write(b, binary.LittleEndian, g.I)
		return
	}
	b.WriteByte(0)
	_ = binary.Write(b, binary.LittleEndian, g.V)
}

func getRow(r io.Reader, ver uint32, n int) (GridRow, error) {
	var g GridRow
	g.Band, _ = getStr(r)
	g.Name, _ = getStr(r)
	if ver >= typedVersion {
		var k [1]byte
		if _, err := io.ReadFull(r, k[:]); err != nil {
			return g, err
		}
		g.Int = k[0] == 1
	}
	g = newRow(g.Band, g.Name, g.Int, n)
	if g.Int {
		return g, binary.Read(r, binary.LittleEndian, g.I)
	}
	return g, binary.Read(r, binary.LittleEndian, g.V)
}
