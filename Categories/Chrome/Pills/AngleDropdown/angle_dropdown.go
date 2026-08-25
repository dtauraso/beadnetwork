package AngleDropdown

import (
	"fmt"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

const (
	Label = "Angles"

	LatticeName = "Lattice points"
	AxisName    = "phi"

	LatticePointsMin  = 4
	LatticePointsMax  = 64
	LatticePointsStep = 4
)

type Rect = Panel.Rect

type Stepper struct {
	Row  Rect
	Name string

	Shown    string
	ValueRow int32
	Denom    int32

	Widest string

	Up   Rect
	Down Rect

	UpEnabled   bool
	DownEnabled bool
}

type Group struct {
	NodeRow int32
	Heading string

	Head Rect
	Open bool

	Phi Stepper
}

type Layout struct {
	Pill    Rect
	Caret   Rect
	Open    bool
	Popover Rect

	Lattice Stepper
	Groups  []Group
}

func AngleDenom(points int32) int32 { return polarindex.AngleDenom(points) }

func WidestAngle(points int32) string {
	if points < 1 {
		points = 1
	}
	return polarindex.AngleText(-points, points)
}

type Node struct {
	Row   int32
	Label string
	Open  bool
}

const arrowW = 18

func stepper(x, y, w float32, name, shown string, valueRow, denom int32, widest string, upOn, downOn bool) Stepper {
	h := Panel.StepperH()
	arrowH := Panel.LineHeight(Panel.PillGlyphPx) + 2*2
	arrowY := y + h - Panel.RowPadY - arrowH
	return Stepper{
		Row:         Rect{X: x, Y: y, W: w, H: h},
		Name:        name,
		Shown:       shown,
		ValueRow:    valueRow,
		Denom:       denom,
		Widest:      widest,
		Down:        Rect{X: x + w - Panel.RowPadX - arrowW, Y: arrowY, W: arrowW, H: arrowH},
		Up:          Rect{X: x + w - Panel.RowPadX - 2*arrowW - 4, Y: arrowY, W: arrowW, H: arrowH},
		UpEnabled:   upOn,
		DownEnabled: downOn,
	}
}

func Build(st *Panel.PillStack, open bool, latticePoints int32, nodes []Node) Layout {
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

	contentH := Panel.StepperH()
	for _, n := range nodes {
		contentH += Panel.HeadingH()
		if n.Open {
			contentH += Panel.StepperH()
		}
	}

	box, x, y := st.AddPopover(contentH)
	lay.Popover = box
	w := box.W - 2*Panel.PopoverPad

	lay.Lattice = stepper(
		x, y, w, LatticeName,
		fmt.Sprintf("%d", latticePoints), -1, 0, fmt.Sprintf("%d", LatticePointsMax),
		latticePoints < LatticePointsMax, latticePoints > LatticePointsMin,
	)
	y += Panel.StepperH()

	lay.Groups = make([]Group, len(nodes))
	for i, n := range nodes {
		heading := n.Label
		if heading == "" {
			heading = fmt.Sprintf("%d", n.Row)
		}
		g := Group{
			NodeRow: n.Row,
			Heading: heading,
			Head:    Rect{X: x, Y: y, W: w, H: Panel.HeadingH()},
			Open:    n.Open,
		}
		y += Panel.HeadingH()
		if n.Open {
			g.Phi = stepper(
				x, y, w, AxisName,
				"", n.Row, AngleDenom(latticePoints), WidestAngle(latticePoints),
				true, true,
			)
			y += Panel.StepperH()
		}
		lay.Groups[i] = g
	}
	st.EndGroup()
	return lay
}

type HitKind int

const (
	HitNone HitKind = iota
	HitPill
	HitGroup
	HitLatticeUp
	HitLatticeDown
	HitPhiUp
	HitPhiDown
)

type Hit struct {
	Kind    HitKind
	NodeRow int32

	Rect Rect
}

func (l Layout) Hit(x, y float64) Hit {
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
	for _, g := range l.Groups {
		if Panel.HitRect(g.Head, x, y) {
			return Hit{Kind: HitGroup, NodeRow: g.NodeRow, Rect: g.Head}
		}
		if !g.Open {
			continue
		}
		if Panel.HitRect(g.Phi.Up, x, y) {
			return Hit{Kind: HitPhiUp, NodeRow: g.NodeRow, Rect: g.Phi.Up}
		}
		if Panel.HitRect(g.Phi.Down, x, y) {
			return Hit{Kind: HitPhiDown, NodeRow: g.NodeRow, Rect: g.Phi.Down}
		}
	}
	return Hit{}
}

type State struct {
	w *ValueWriter // this piece's own writer, armed when the scene opens

	Open      bool
	GroupOpen map[int32]bool
}

func (s *State) Arm(sceneRoot string) { s.w = NewValueWriter(sceneRoot) }
