package viewstate

import (
	"encoding/binary"
	"math"

	"github.com/dtauraso/wirefold/src/Node/Wiring/angledropdown"
	B "github.com/dtauraso/wirefold/src/Buffer"
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

func (r *runCols) Rect(xc, yc, wc, hc int, rect angledropdown.Rect) {
	r.F32(xc, rect.X)
	r.F32(yc, rect.Y)
	r.F32(wc, rect.W)
	r.F32(hc, rect.H)
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

func (ui *UIState) writeAnglePillColumns(lay angledropdown.Layout) {
	c := ui.singletonCols
	if c == nil {
		return
	}

	c.SetF32(B.ColStreamAnglePillPillX, lay.Pill.X)
	c.SetF32(B.ColStreamAnglePillPillY, lay.Pill.Y)
	c.SetF32(B.ColStreamAnglePillPillW, lay.Pill.W)
	c.SetF32(B.ColStreamAnglePillPillH, lay.Pill.H)
	c.SetU8(B.ColStreamAnglePillOpen, boolU8(lay.Open))
	c.SetF32(B.ColStreamAnglePillPopoverX, lay.Popover.X)
	c.SetF32(B.ColStreamAnglePillPopoverY, lay.Popover.Y)
	c.SetF32(B.ColStreamAnglePillPopoverW, lay.Popover.W)
	c.SetF32(B.ColStreamAnglePillPopoverH, lay.Popover.H)
	c.SetBytes(B.ColStreamAnglePillLabelText, []byte(angledropdown.Label))

	steps := newRunCols()
	addStep := func(s angledropdown.Stepper) {
		steps.Rect(B.ColStreamAnglePillStepX, B.ColStreamAnglePillStepY, B.ColStreamAnglePillStepW, B.ColStreamAnglePillStepH, s.Row)
		steps.Str(B.ColStreamAnglePillStepNameText, B.ColStreamAnglePillStepNameLen, s.Name)
		steps.Str(B.ColStreamAnglePillStepShownText, B.ColStreamAnglePillStepShownLen, s.Shown)
		steps.I32(B.ColStreamAnglePillStepValueRow, s.ValueRow)
		steps.I32(B.ColStreamAnglePillStepDenom, s.Denom)
		steps.Rect(B.ColStreamAnglePillStepUpX, B.ColStreamAnglePillStepUpY, B.ColStreamAnglePillStepUpW, B.ColStreamAnglePillStepUpH, s.Up)
		steps.Rect(B.ColStreamAnglePillStepDownX, B.ColStreamAnglePillStepDownY, B.ColStreamAnglePillStepDownW, B.ColStreamAnglePillStepDownH, s.Down)
		steps.U8(B.ColStreamAnglePillStepUpOn, boolU8(s.UpEnabled))
		steps.U8(B.ColStreamAnglePillStepDownOn, boolU8(s.DownEnabled))
	}

	groups := newRunCols()

	if lay.Open {
		addStep(lay.Lattice)
		for _, g := range lay.Groups {
			groups.Rect(B.ColStreamAnglePillGroupX, B.ColStreamAnglePillGroupY, B.ColStreamAnglePillGroupW, B.ColStreamAnglePillGroupH, g.Head)
			groups.U8(B.ColStreamAnglePillGroupOpen, boolU8(g.Open))
			groups.Str(B.ColStreamAnglePillGroupHeadText, B.ColStreamAnglePillGroupHeadLen, g.Heading)
			if g.Open {
				addStep(g.Phi)
			}
		}
	}

	steps.writeTo(c,
		B.ColStreamAnglePillStepX, B.ColStreamAnglePillStepY, B.ColStreamAnglePillStepW, B.ColStreamAnglePillStepH,
		B.ColStreamAnglePillStepNameText, B.ColStreamAnglePillStepNameLen,
		B.ColStreamAnglePillStepShownText, B.ColStreamAnglePillStepShownLen,
		B.ColStreamAnglePillStepValueRow, B.ColStreamAnglePillStepDenom,
		B.ColStreamAnglePillStepUpX, B.ColStreamAnglePillStepUpY, B.ColStreamAnglePillStepUpW, B.ColStreamAnglePillStepUpH,
		B.ColStreamAnglePillStepDownX, B.ColStreamAnglePillStepDownY, B.ColStreamAnglePillStepDownW, B.ColStreamAnglePillStepDownH,
		B.ColStreamAnglePillStepUpOn, B.ColStreamAnglePillStepDownOn,
	)
	groups.writeTo(c,
		B.ColStreamAnglePillGroupX, B.ColStreamAnglePillGroupY, B.ColStreamAnglePillGroupW, B.ColStreamAnglePillGroupH,
		B.ColStreamAnglePillGroupOpen,
		B.ColStreamAnglePillGroupHeadText, B.ColStreamAnglePillGroupHeadLen,
	)
}
