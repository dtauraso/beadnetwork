package CardPanel

import (
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
)

const (
	TitleFontPx = 12
	KeyFontPx   = 11
	ValFontPx   = 13

	CellPadX = 6
	CellPadY = 2
	CellGapX = 6
	FieldGap = 4

	TitleGap = 3
	ColGap   = 8
	TopY     = 44

	valueField = "-0000"
)

type Rect = Panel.Rect

type PanelBox struct {
	Box   Rect
	Title string
	Head  Rect
}

type FieldBox struct {
	Field Field
	Key   string
	Rect  Rect
}

type Layout struct {
	Panels []PanelBox
	Fields []FieldBox
}

type Edit struct {
	Active bool
	Field  Field
	Draft  string
}

type State struct {
	w *ValueWriter

	Nodes []int
	Cards map[int]Card
	S     int
	Edit  Edit
}

func (s *State) Arm(sceneRoot string) { s.w = NewValueWriter(sceneRoot) }

func (s State) Value(f Field) int {
	if f.Vector == VecS {
		return s.S
	}
	return s.Cards[f.Node].Get(f)
}

var subscripts = []string{"₀", "₁", "₂", "₃", "₄", "₅", "₆", "₇", "₈", "₉"}

func sub(n int) string { return subscripts[n%10] }

var compKeys = map[Comp]string{CompPhi: "φ", CompTheta: "θ", CompR: "r"}

func keyOf(f Field) string {
	switch f.Vector {
	case VecK:
		return "k" + sub(f.J)
	case VecL:
		return "L" + sub(f.J)
	case VecS:
		return "s"
	}
	return compKeys[f.Comp]
}

func titleOf(node int, f Field) string {
	head := "node " + strconv.Itoa(node) + " · "
	switch f.Vector {
	case VecStart:
		return head + "start" + sub(f.J)
	case VecPoleOffset:
		return head + "pole_offset" + sub(f.J)
	case VecK:
		return head + "k"
	case VecL:
		return head + "L"
	}
	return "s"
}

type group struct {
	title  string
	fields []Field
}

func groupsOf(node int) []group {
	var out []group
	for _, f := range NodeFields(node) {
		t := titleOf(node, f)
		if len(out) == 0 || out[len(out)-1].title != t {
			out = append(out, group{title: t})
		}
		out[len(out)-1].fields = append(out[len(out)-1].fields, f)
	}
	return out
}

func fieldW(key string) float32 {
	return Panel.TextWidth(key, KeyFontPx) + CellGapX + Panel.TextWidth(valueField, ValFontPx) + 2*CellPadX
}

func groupW(g group) float32 {
	var w float32
	for i, f := range g.fields {
		if i > 0 {
			w += FieldGap
		}
		w += fieldW(keyOf(f))
	}
	if t := Panel.TextWidth(g.title, TitleFontPx); t > w {
		return t
	}
	return w
}

func groupH() float32 {
	return Panel.LineHeight(TitleFontPx) + TitleGap + Panel.LineHeight(ValFontPx) + 2*CellPadY
}

func (lay *Layout) place(g group, box Rect, x, y float32) {
	headH := Panel.LineHeight(TitleFontPx)
	lay.Panels = append(lay.Panels, PanelBox{Box: box, Title: g.title, Head: Rect{X: x, Y: y, W: box.W - 2*Panel.PadX, H: headH}})
	fy := y + headH + TitleGap
	fh := Panel.LineHeight(ValFontPx) + 2*CellPadY
	fx := x
	for _, f := range g.fields {
		key := keyOf(f)
		w := fieldW(key)
		lay.Fields = append(lay.Fields, FieldBox{Field: f, Key: key, Rect: Rect{X: fx, Y: fy, W: w, H: fh}})
		fx += w + FieldGap
	}
}

func Build(st *Panel.Stack, viewW float32, s State) Layout {
	if len(s.Nodes) == 0 {
		return Layout{}
	}
	var lay Layout

	sGroup := group{title: "s", fields: []Field{{Vector: VecS, Comp: CompOne}}}
	box, x, y := st.Add(groupW(sGroup), groupH())
	lay.place(sGroup, box, x, y)

	cols := make([][]group, len(s.Nodes))
	widths := make([]float32, len(s.Nodes))
	var total float32
	for i, node := range s.Nodes {
		cols[i] = groupsOf(node)
		for _, g := range cols[i] {
			if w := groupW(g); w > widths[i] {
				widths[i] = w
			}
		}
		total += widths[i] + 2*Panel.PadX
		if i > 0 {
			total += ColGap
		}
	}

	cx := viewW - Panel.OriginX - total
	for i := range s.Nodes {
		cy := float32(TopY)
		for _, g := range cols[i] {
			box := Rect{X: cx, Y: cy, W: widths[i] + 2*Panel.PadX, H: groupH() + 2*Panel.PadY}
			lay.place(g, box, cx+Panel.PadX, cy+Panel.PadY)
			cy += box.H + Panel.Gap
		}
		cx += widths[i] + 2*Panel.PadX + ColGap
	}
	return lay
}

func (l Layout) Hit(x, y float64) (Field, bool) {
	for _, fb := range l.Fields {
		if Panel.HitRect(fb.Rect, x, y) {
			return fb.Field, true
		}
	}
	return Field{}, false
}
