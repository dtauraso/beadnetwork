package Sheet

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
)

const (
	TitleFontPx = 12
	KeyFontPx   = 11
	ValFontPx   = 13

	CellPadX = 6
	CellPadY = 2
	CellGapX = 6
	RowGapY  = 3
	TitleGap = 3

	StepArrowW = 14
	StepGap    = 2

	LockGap = 8

	IndentX = 18

	valueField = "-00.00"
)

type Rect = Panel.Rect

type CellKind uint8

const (
	CellText CellKind = iota
	CellField
	CellButton
)

type Act struct {
	Op   string
	Node string
	To   string
	K    string
	Sign int
	I    int
}

type Cell struct {
	Kind          CellKind
	Key           string
	Value         string
	Act           Act
	Typed         bool
	Steps         bool
	Lit           bool
	Fit           string
	OverLastDigit bool
	Boxed         bool
	Outlined      bool
	Arrows        Act
	LockSign      int
	Locked        bool
	Dim           bool
	Words         bool
}

func (c Cell) LockedOn(on bool) Cell { c.Locked = on; return c }

func (c Cell) Dimmed(on bool) Cell { c.Dim = on; return c }

func (c Cell) Wording() Cell { c.Words = true; return c }

func (c Cell) Stepping(a Act, lock int) Cell { c.Arrows, c.LockSign = a, lock; return c }

func (c Cell) ArrowAct() Act {
	if c.Arrows.Op != "" {
		return c.Arrows
	}
	return c.Act
}

func (c Cell) Fitting(v string) Cell { c.Fit = v; return c }

func (c Cell) Outline(on bool) Cell { c.Boxed, c.Outlined = true, on; return c }

func (c Cell) Acting(a Act) Cell { c.Act = a; return c }

func (c Cell) AboveLastDigit() Cell { c.OverLastDigit = true; return c }

func Text(s string) Cell { return Cell{Kind: CellText, Key: s} }

func LitText(s string, act Act, lit bool) Cell {
	return Cell{Kind: CellText, Key: s, Act: act, Lit: lit}
}

func Button(label string, act Act, lit bool) Cell {
	return Cell{Kind: CellButton, Key: label, Act: act, Lit: lit}
}

func Field(key, value string, act Act, typed, steps bool) Cell {
	return Cell{Kind: CellField, Key: key, Value: value, Act: act, Typed: typed, Steps: steps}
}

type Group struct {
	Title   string
	Fold    Act
	Rows    [][]Cell
	Slides  bool
	Table   string
	Wraps   bool
	Indent  bool
	Card    bool
	Pop     *Anchor
	MaxRows int
}

type Anchor struct {
	Group, Col int
}

type SlideBox struct {
	Table string
	Rect  Rect
	Max   float32
}

type Pill struct {
	Label  string
	Left   bool
	Bottom bool
	Groups []Group
}

type PillBox struct {
	Label     string
	Left      bool
	Pill      Rect
	Open      bool
	Popover   Rect
	Grab      Rect
	Bar       bool
	Tabs      []DockTab
	MaxScroll float32
	Slide     Rect
	Slides    []SlideBox
}

type DockTab struct {
	Label string
	Rect  Rect
	On    bool
}

type HeadBox struct {
	Box   int
	Rect  Rect
	Title string
	Fold  Act
}

type CellBox struct {
	Box    int
	Cell   Cell
	Rect   Rect
	Up     Rect
	Down   Rect
	Slides bool
}

type Layout struct {
	Pills []PillBox
	Heads []HeadBox
	Cells []CellBox
	Cards []CardBox
}
