package rulespanel

import "github.com/dtauraso/wirefold/nodes/Wiring/panelstack"

const (
	Label       = "polar rules"
	LabelClosed = "▸ polar rules"
	LabelOpen   = "▾ polar rules"

	OriginX = 8

	PadX = 10
	PadY = 6

	Width = 260

	FontPx     = 12
	HeadFontPx = 11

	RowPadY    = 1
	BlockGap   = 4
	IndentName = 10
	IndentLine = 10

	CheckSize = 13
	NameW     = 78
	GlyphW    = 12
	Gap       = 6
)

type Rect = panelstack.Rect

type RowKind int

const (
	RowToggle RowKind = iota
	RowNodeHead
	RowHolderHead
	RowLine
	RowText
)

type ValueKind int

const (
	ValNone ValueKind = iota
	ValFree

	ValSelfR
	ValSelfPhi
	ValSelfTheta

	ValDragR
	ValDragPhi
	ValDragTheta
)

type CheckKind int

const (
	CheckNone CheckKind = iota
	CheckNodeDrag
	CheckSelfDrag
	CheckEdgeDrag
	CheckKindRule
)

type Row struct {
	Kind  RowKind
	Rect  Rect
	Depth int

	Text  string
	Glyph string

	Free bool

	NodeRow int32
	EdgeRow int32

	Check     CheckKind
	CheckRect Rect

	Value     ValueKind
	ValueRect Rect

	SharedRect Rect

	Editing bool
}

type MenuRow struct {
	Rect      Rect
	CheckRect Rect
	Label     string

	NodeRow int32
}

type Layout struct {
	Box  Rect
	Open bool

	Toggle Rect

	Rows []Row

	Draft     string
	DraftRect Rect

	MenuOpen      bool
	MenuAnchorRow int32
	MenuBox       Rect
	MenuRows      []MenuRow
}

type Node struct {
	Row   int32
	Label string
	Kind  string

	Out []Edge

	HasReverse []bool
}

type Edge struct {
	EdgeRow    int32
	OtherRow   int32
	OtherLabel string
}

type Edit struct {
	Active  bool
	NodeRow int32
	Self    bool
	Draft   string
}

func Build(st *panelstack.Stack, open bool, nodes []Node, edit Edit, sharedMenuRow int32) Layout {
	toggleH := panelstack.LineHeight(HeadFontPx) + 4
	top := st.Next()
	lay := Layout{
		Open:   open,
		Toggle: Rect{X: OriginX + PadX, Y: top + PadY, W: 90, H: toggleH},
	}

	b := &builder{
		x:    OriginX + PadX,
		w:    Width - 2*PadX,
		y:    top + PadY + toggleH,
		edit: edit,
	}

	if open {
		for _, n := range nodes {
			b.y += BlockGap
			buildNode(b, n)
		}
	}

	lay.Rows = b.rows
	lay.Box = Rect{X: OriginX, Y: top, W: Width, H: b.y - top + PadY}
	st.Took(lay.Box.H)
	if edit.Active {
		lay.Draft = edit.Draft
		for _, r := range b.rows {
			if r.Editing {
				lay.DraftRect = r.ValueRect
			}
		}
	}
	if open && sharedMenuRow >= 0 {
		buildSharedMenu(&lay, nodes, sharedMenuRow)
	}
	return lay
}
