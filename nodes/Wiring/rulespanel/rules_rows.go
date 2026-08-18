package rulespanel

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/nodedrag"
	"github.com/dtauraso/wirefold/nodes/Wiring/panelstack"
)

type builder struct {
	rows []Row
	y    float32
	x    float32
	w    float32
	edit Edit
}

func (b *builder) line(depth int, glyph, text string, free bool, node, edge int32, v ValueKind, editing bool) {
	h := panelstack.LineHeight(FontPx) + 2*RowPadY
	x := b.x + float32(depth)*IndentLine
	r := Row{
		Kind: RowLine, Rect: Rect{X: x, Y: b.y, W: b.w - float32(depth)*IndentLine, H: h},
		Depth: depth, Glyph: glyph, Text: text, Free: free,
		NodeRow: node, EdgeRow: edge, Value: v, Editing: editing,
	}
	r.ValueRect = Rect{X: x + GlyphW + Gap, Y: b.y, W: r.Rect.W - GlyphW - Gap, H: h}
	b.rows = append(b.rows, r)
	b.y += h
}

func (b *builder) holderHead(depth int, text string, check CheckKind, node, edge int32) {
	h := panelstack.LineHeight(FontPx) + 2*RowPadY
	x := b.x + float32(depth)*IndentName
	r := Row{
		Kind: RowHolderHead, Rect: Rect{X: x, Y: b.y, W: b.w - float32(depth)*IndentName, H: h},
		Depth: depth, Text: text, Check: check, NodeRow: node, EdgeRow: edge,
	}
	r.CheckRect = Rect{X: x, Y: b.y + (h-CheckSize)/2, W: CheckSize, H: CheckSize}
	b.rows = append(b.rows, r)
	b.y += h
}

func (b *builder) text(depth int, text string, free bool) {
	h := panelstack.LineHeight(FontPx) + 2*RowPadY
	x := b.x + float32(depth)*IndentLine
	b.rows = append(b.rows, Row{
		Kind: RowText, Rect: Rect{X: x, Y: b.y, W: b.w - float32(depth)*IndentLine, H: h},
		Depth: depth, Text: text, Free: free,
	})
	b.y += h
}

func buildNode(b *builder, n Node) {
	headH := panelstack.LineHeight(FontPx) + 2*RowPadY
	head := Row{
		Kind: RowNodeHead, Rect: Rect{X: b.x, Y: b.y, W: b.w, H: headH},
		Text: n.Label, Glyph: n.Kind, NodeRow: n.Row, Check: CheckNodeDrag,
	}
	head.CheckRect = Rect{X: b.x, Y: b.y + (headH-CheckSize)/2, W: CheckSize, H: CheckSize}
	sharedW := panelstack.TextWidth("⇄ shared ×00", FontPx) + 10
	head.SharedRect = Rect{X: b.x + b.w - sharedW, Y: b.y, W: sharedW, H: headH}
	b.rows = append(b.rows, head)
	b.y += headH

	b.holderHead(1, n.Label+" itself", CheckSelfDrag, n.Row, -1)
	selfEdit := b.edit.Active && b.edit.Self && b.edit.NodeRow == n.Row
	b.line(2, "r", "", false, n.Row, -1, ValSelfR, false)
	b.line(2, "φ", "", false, n.Row, -1, ValSelfPhi, false)
	b.line(2, "θ", "", false, n.Row, -1, ValSelfTheta, selfEdit)

	dragEdit := b.edit.Active && !b.edit.Self && b.edit.NodeRow == n.Row
	for i, e := range n.Out {
		b.holderHead(1, n.Label+" → "+e.OtherLabel, CheckEdgeDrag, n.Row, e.EdgeRow)
		if i < len(n.HasReverse) && n.HasReverse[i] {
			b.line(2, "r", "", false, n.Row, e.EdgeRow, ValDragR, false)
			b.line(2, "φ", "", false, n.Row, -1, ValDragPhi, false)
			b.line(2, "θ", "", false, n.Row, -1, ValDragTheta, dragEdit)
			continue
		}
		b.line(2, "—", "no rule of its own", true, n.Row, -1, ValNone, false)
	}

	if !nodedrag.HasKindRule(n.Kind) {
		return
	}
	count := len(n.Out)
	name := fmt.Sprintf("all %d", count)
	if count == 2 {
		name = "both"
	}
	b.holderHead(1, name, CheckKindRule, n.Row, -1)
	for _, a := range KindAxisRules[n.Kind] {
		b.line(2, a.Glyph, a.Text, false, n.Row, -1, ValNone, false)
	}
	for _, s := range KindSpanningRules[n.Kind] {
		b.text(2, "⤷ "+s, false)
	}
}

func buildSharedMenu(lay *Layout, nodes []Node, anchorRow int32) {
	var anchor Rect
	for _, r := range lay.Rows {
		if r.Kind == RowNodeHead && r.NodeRow == anchorRow {
			anchor = r.SharedRect
		}
	}
	if anchor.W == 0 {
		return
	}

	rowH := panelstack.LineHeight(HeadFontPx) + 4
	headH := rowH
	w := float32(128)
	for _, n := range nodes {
		if t := panelstack.TextWidth(n.Label, HeadFontPx) + CheckSize + 3*Gap; t > w {
			w = t
		}
	}

	lay.MenuOpen = true
	lay.MenuAnchorRow = anchorRow
	x := anchor.X + anchor.W + 8
	y := anchor.Y
	lay.MenuBox = Rect{X: x, Y: y, W: w + 12, H: headH + float32(len(nodes)+1)*rowH + 8}

	rowY := y + headH + 4
	add := func(label string, row int32) {
		r := MenuRow{
			Rect:  Rect{X: x + 6, Y: rowY, W: w, H: rowH},
			Label: label, NodeRow: row,
		}
		r.CheckRect = Rect{X: x + 6, Y: rowY + (rowH-CheckSize)/2, W: CheckSize, H: CheckSize}
		lay.MenuRows = append(lay.MenuRows, r)
		rowY += rowH
	}
	add("all nodes", -1)
	for _, n := range nodes {
		add(n.Label, n.Row)
	}
}

var KindAxisRules = map[string][]struct{ Glyph, Text string }{
	"Input": {{Glyph: "θ", Text: "snaps to half-turns"}},
}

var KindSpanningRules = map[string][]string{
	"Input": {"out-lengths held equal"},
}
