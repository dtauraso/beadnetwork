package Store

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Files"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	FileGrid          = "grid.bin"
	gridMagic         = "GRID"
	gridVersion       = 5
	gapOriginsVersion = 4
	oneColGridVersion = 6
)

func gridVersionOf(path string) int {
	data, err := os.ReadFile(path)
	if err != nil || len(data) < len(gridMagic)+4 || !strings.HasPrefix(string(data), gridMagic) {
		return 0
	}
	return int(binary.LittleEndian.Uint32(data[len(gridMagic):]))
}

func IsOneColGrid(path string) bool { return gridVersionOf(path) == oneColGridVersion }

type OldBase struct {
	Origin Links.Start
	Total  Links.Ints
}

type OldGrid struct {
	ID, Level string
	Shown     int
	Cols      []Col
	Base      map[string]OldBase
}

func (g OldGrid) ShownCol() Col {
	for _, c := range g.Cols {
		if c.Step == g.Shown {
			return c
		}
	}
	if len(g.Cols) > 0 {
		return g.Cols[0]
	}
	return Col{}
}

func GridPath(nodeDir string) string { return filepath.Join(nodeDir, FileGrid) }

func OldGrids(roots []string) []string {
	var out []string
	for _, root := range roots {
		found, _ := filepath.Glob(GridPath(Files.NodeDir(root, "*")))
		out = append(out, found...)
	}
	return out
}

type GridHead struct {
	ID, Level string
	Shown     int
	Ver       int
	Steps     []int
	Base      []GridRow
}

func ReadGridRows(path string) (GridHead, []GridRow, error) {
	var h GridHead
	data, err := os.ReadFile(path)
	if err != nil {
		return h, nil, err
	}
	if !strings.HasPrefix(string(data), gridMagic) {
		return h, nil, fmt.Errorf("Store.ReadGridRows: %s does not start with %q, so it is not a node grid", path, gridMagic)
	}
	r, le := bytes.NewReader(data[len(gridMagic):]), binary.LittleEndian
	var ver, ncols, nrows uint32
	var shown int64
	_ = binary.Read(r, le, &ver)
	if ver < 1 || ver > gridVersion {
		return h, nil, fmt.Errorf("Store.ReadGridRows: %s is grid version %d, and this build reads versions 1 to %d", path, ver, gridVersion)
	}
	h.Ver = int(ver)
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
		return h, nil, fmt.Errorf("Store.ReadGridRows: %s ends inside its header: %w", path, err)
	}
	rows := make([]GridRow, nrows)
	for i := range rows {
		if rows[i], err = getRow(r, ver, int(ncols)); err != nil {
			return h, nil, fmt.Errorf("Store.ReadGridRows: %s ends inside row %d (%s %s): %w", path, i, rows[i].Band, rows[i].Name, err)
		}
	}
	if ver >= 2 {
		if h.Base, err = readBase(r, ver); err != nil {
			return h, nil, fmt.Errorf("Store.ReadGridRows: %s ends inside its running totals: %w", path, err)
		}
	}
	return h, rows, nil
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

func ReadOldGrid(path string) (OldGrid, bool, error) {
	h, rows, err := ReadGridRows(path)
	if errors.Is(err, fs.ErrNotExist) {
		return OldGrid{}, false, nil
	}
	if err != nil {
		return OldGrid{}, false, err
	}
	g := OldGrid{ID: h.ID, Level: h.Level, Shown: h.Shown, Cols: make([]Col, len(h.Steps)), Base: map[string]OldBase{}}
	for i, s := range h.Steps {
		g.Cols[i].Step = s
	}
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
	if err := colsOf(h.ID, rows, g.Cols); err != nil {
		return g, false, fmt.Errorf("Store.ReadOldGrid: %s %w", path, err)
	}
	if h.Ver < gapOriginsVersion {
		gapOrigins(g)
	}
	return g, true, nil
}

func gapOrigins(g OldGrid) {
	shown := g.ShownCol()
	for from, b := range g.Base {
		in, ok := shown.InFrom(from)
		if !ok || in.Link.Picked {
			continue
		}
		b.Origin.Less(b.Total.Add(in.Link.Delta))
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
