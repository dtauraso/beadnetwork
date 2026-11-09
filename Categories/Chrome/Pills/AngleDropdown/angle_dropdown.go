package AngleDropdown

import (
	"fmt"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
)

const (
	Label = "Angles"

	LatticeName = "Lattice points"

	LatticePointsMin  = 4
	LatticePointsMax  = 64
	LatticePointsStep = 4
)

type Rect = Panel.Rect

type Stepper struct {
	Row  Rect
	Name string

	Shown string

	Up   Rect
	Down Rect

	UpEnabled   bool
	DownEnabled bool
}

type Layout struct {
	Pill    Rect
	Caret   Rect
	Open    bool
	Popover Rect

	Lattice Stepper
}

const arrowW = 18

func stepper(x, y, w float32, name, shown string, upOn, downOn bool) Stepper {
	h := Panel.StepperH()
	arrowH := Panel.LineHeight(Panel.PillGlyphPx) + 2*2
	arrowY := y + h - Panel.RowPadY - arrowH
	return Stepper{
		Row:         Rect{X: x, Y: y, W: w, H: h},
		Name:        name,
		Shown:       shown,
		Down:        Rect{X: x + w - Panel.RowPadX - arrowW, Y: arrowY, W: arrowW, H: arrowH},
		Up:          Rect{X: x + w - Panel.RowPadX - 2*arrowW - 4, Y: arrowY, W: arrowW, H: arrowH},
		UpEnabled:   upOn,
		DownEnabled: downOn,
	}
}

func Build(st *Panel.PillStack, open bool, latticePoints int32) Layout {
	pill := st.AddPill()
	lay := Layout{
		Pill:  pill,
		Caret: Rect{X: pill.X + pill.W - Panel.CaretW, Y: pill.Y, W: Panel.CaretW, H: pill.H},
		Open:  open,
	}
	if !open {
		st.EndGroup()
		return lay
	}

	box, x, y := st.AddPopover(Panel.StepperH())
	lay.Popover = box
	w := box.W - 2*Panel.PopoverPad

	lay.Lattice = stepper(
		x, y, w, LatticeName,
		fmt.Sprintf("%d", latticePoints),
		latticePoints < LatticePointsMax, latticePoints > LatticePointsMin,
	)
	st.EndGroup()
	return lay
}

type HitKind int

const (
	HitNone HitKind = iota
	HitPill
	HitLatticeUp
	HitLatticeDown
)

type Hit struct {
	Kind HitKind

	Rect Rect
}

func (l Layout) Hit(x, y float64) Hit {
	if l.Pill.W <= 0 {
		return Hit{}
	}
	if Panel.HitRect(l.Pill, x, y) {
		return Hit{Kind: HitPill, Rect: l.Pill}
	}
	if !l.Open {
		return Hit{}
	}
	if l.Lattice.UpEnabled && Panel.HitRect(l.Lattice.Up, x, y) {
		return Hit{Kind: HitLatticeUp, Rect: l.Lattice.Up}
	}
	if l.Lattice.DownEnabled && Panel.HitRect(l.Lattice.Down, x, y) {
		return Hit{Kind: HitLatticeDown, Rect: l.Lattice.Down}
	}
	return Hit{}
}

type State struct {
	w *ValueWriter // this piece's own writer, armed when the scene opens

	Open bool
}

func (s *State) Arm(sceneRoot string) { s.w = NewValueWriter(sceneRoot) }
