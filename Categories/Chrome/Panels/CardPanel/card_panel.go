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

const PillLabel = "Node vectors"

type Layout struct {
	Pill      Rect
	Open      bool
	Popover   Rect
	MaxScroll float32

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

	Nodes  []int
	Cards  map[int]Card
	S      int
	Edit   Edit
	Open   bool
	Scroll float32
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
	lay.Panels = append(lay.Panels, PanelBox{Box: box, Title: g.title, Head: Rect{X: x, Y: y, W: groupW(g), H: headH}})
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

func Build(pills *Panel.PillStack, s State) Layout {
	if len(s.Nodes) == 0 {
		return Layout{}
	}
	lay := Layout{Pill: pills.AddPill(), Open: s.Open}
	if !s.Open {
		pills.EndGroup()
		return lay
	}

	groups := []group{{title: "s", fields: []Field{{Vector: VecS, Comp: CompOne}}}}
	for _, node := range s.Nodes {
		groups = append(groups, groupsOf(node)...)
	}
	var contentW float32
	for _, g := range groups {
		if w := groupW(g); w > contentW {
			contentW = w
		}
	}
	contentH := float32(len(groups))*(groupH()+Panel.Gap) - Panel.Gap

	box, cx, cy, maxScroll := pills.AddWideScrollingPopover(contentW, contentH, s.Scroll)
	pills.EndGroup()
	lay.Popover, lay.MaxScroll = box, maxScroll
	lay.Panels = append(lay.Panels, PanelBox{Box: box})
	for _, g := range groups {
		lay.place(g, Rect{}, cx, cy)
		cy += groupH() + Panel.Gap
	}
	return lay
}

func (l Layout) HitPill(x, y float64) bool {
	return l.Pill.W > 0 && Panel.HitRect(l.Pill, x, y)
}

func (l Layout) Covers(x, y float64) bool {
	return l.Open && Panel.HitRect(l.Popover, x, y)
}

func (l Layout) Hit(x, y float64) (Field, bool) {
	if !l.Covers(x, y) {
		return Field{}, false
	}
	for _, fb := range l.Fields {
		if Panel.HitRect(fb.Rect, x, y) {
			return fb.Field, true
		}
	}
	return Field{}, false
}
