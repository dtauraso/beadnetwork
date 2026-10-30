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
	Outlined      bool
}

func (c Cell) Fitting(v string) Cell { c.Fit = v; return c }

func (c Cell) Outline(on bool) Cell { c.Outlined = on; return c }

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
	Title string
	Rows  [][]Cell
}

type Pill struct {
	Label  string
	Left   bool
	Groups []Group
}

type PillBox struct {
	Label     string
	Left      bool
	Pill      Rect
	Open      bool
	Popover   Rect
	MaxScroll float32
}

type HeadBox struct {
	Box   int
	Rect  Rect
	Title string
}

type CellBox struct {
	Box  int
	Cell Cell
	Rect Rect
	Up   Rect
	Down Rect
}

type Layout struct {
	Pills []PillBox
	Heads []HeadBox
	Cells []CellBox
}
