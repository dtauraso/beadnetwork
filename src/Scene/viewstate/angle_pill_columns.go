package viewstate

import (
	"encoding/binary"
	"math"

	"github.com/dtauraso/wirefold/src/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/src/valuefile"
)

type runCols struct {
	f32  map[int][]byte
	u8   map[int][]byte
	i32  map[int][]byte
	u32  map[int][]byte
	text map[int][]byte
}

func newRunCols() *runCols {
	return &runCols{
		f32: map[int][]byte{}, u8: map[int][]byte{},
		i32: map[int][]byte{}, u32: map[int][]byte{}, text: map[int][]byte{},
	}
}

func (r *runCols) F32(col int, v float32) {
	r.f32[col] = binary.LittleEndian.AppendUint32(r.f32[col], math.Float32bits(v))
}

func (r *runCols) U8(col int, v uint8) { r.u8[col] = append(r.u8[col], v) }

func (r *runCols) I32(col int, v int32) {
	r.i32[col] = binary.LittleEndian.AppendUint32(r.i32[col], uint32(v))
}

func (r *runCols) U32(col int, v uint32) {
	r.u32[col] = binary.LittleEndian.AppendUint32(r.u32[col], v)
}

func (r *runCols) Str(textCol, lenCol int, s string) {
	r.text[textCol] = append(r.text[textCol], s...)
	r.u32[lenCol] = binary.LittleEndian.AppendUint32(r.u32[lenCol], uint32(len(s)))
}

func (r *runCols) Rect(xc, yc, wc, hc int, rect AngleDropdown.Rect) {
	r.F32(xc, rect.X)
	r.F32(yc, rect.Y)
	r.F32(wc, rect.W)
	r.F32(hc, rect.H)
}

func (r *runCols) Point(xc, yc int, rect AngleDropdown.Rect) {
	r.F32(xc, rect.X)
	r.F32(yc, rect.Y)
}

func (r *runCols) writeTo(set interface{ SetBytes(int, []byte) }, cols ...int) {
	for _, c := range cols {
		switch {
		case r.f32[c] != nil:
			set.SetBytes(c, r.f32[c])
		case r.u8[c] != nil:
			set.SetBytes(c, r.u8[c])
		case r.i32[c] != nil:
			set.SetBytes(c, r.i32[c])
		case r.u32[c] != nil:
			set.SetBytes(c, r.u32[c])
		case r.text[c] != nil:
			set.SetBytes(c, r.text[c])
		default:
			set.SetBytes(c, nil)
		}
	}
}

func (ui *UIState) writeAnglePillValues(lay AngleDropdown.Layout) {
	w := ui.anglePillValues
	if w == nil {
		return
	}
	w.Begin()

	w.Rect("pillX", "pillY", "pillW", "pillH", lay.Pill)
	w.Bool("open", lay.Open)
	w.Rect("popoverX", "popoverY", "popoverW", "popoverH", lay.Popover)
	w.Text("labelText", AngleDropdown.Label)

	addStep := func(s AngleDropdown.Stepper) {
		w.Rect("stepX", "stepY", "stepW", "stepH", s.Row)
		w.Str("stepNameText", "stepNameLen", s.Name)
		w.Str("stepShownText", "stepShownLen", s.Shown)
		w.I32("stepValueRow", s.ValueRow)
		w.I32("stepDenom", s.Denom)
		w.Rect("stepUpX", "stepUpY", "stepUpW", "stepUpH", s.Up)
		w.Rect("stepDownX", "stepDownY", "stepDownW", "stepDownH", s.Down)
		w.Bool("stepUpOn", s.UpEnabled)
		w.Bool("stepDownOn", s.DownEnabled)
	}

	if lay.Open {
		addStep(lay.Lattice)
		for _, g := range lay.Groups {
			w.Rect("groupX", "groupY", "groupW", "groupH", g.Head)
			w.Bool("groupOpen", g.Open)
			w.Str("groupHeadText", "groupHeadLen", g.Heading)
			if g.Open {
				addStep(g.Phi)
			}
		}
	}

	if err := w.Flush(); err != nil {
		valuefile.LogPersistErr("angle_pill_values", "", err)
	}
}
