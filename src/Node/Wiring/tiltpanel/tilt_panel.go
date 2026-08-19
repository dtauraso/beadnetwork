package tiltpanel

import "github.com/dtauraso/wirefold/src/Node/Wiring/panelstack"

const (
	StartLabel = "start tilt"
	ResetLabel = "reset tilt"

	KeyRounds = "rounds"
	KeyMsgs   = "msgs"
)

const (
	ButtonFontPx = 12
	ButtonPadX   = 10
	ButtonPadY   = 3
	ButtonGap    = 6

	HeadFontPx = 13
	ValFontPx  = 13
	KeyFontPx  = 11

	CellPadX = 7
	CellPadY = 2
	CellGapX = 8

	ColGap = 6
	RowGap = 3

	SectionGap = 6
)

type Rect = panelstack.Rect

type Column struct {
	NodeRow int32
	Label   string

	Head   Rect
	Rounds Rect
	Msgs   Rect
}

type Layout struct {
	Box   Rect
	Start Rect
	Reset Rect

	Columns []Column
}

func buttonSize(label string) (w, h float32) {
	return panelstack.TextWidth(label, ButtonFontPx) + 2*ButtonPadX + 2,
		panelstack.LineHeight(ButtonFontPx) + 2*ButtonPadY + 2
}

const valueField = "000"

func colWidth(label string) float32 {
	head := panelstack.TextWidth("node "+label, HeadFontPx)
	cell := panelstack.TextWidth(KeyRounds, KeyFontPx) + CellGapX +
		panelstack.TextWidth(valueField, ValFontPx) + 2*CellPadX
	if head > cell {
		return head
	}
	return cell
}

func Build(st *panelstack.Stack, rows []int32, labels []string) Layout {
	startW, btnH := buttonSize(StartLabel)
	resetW, _ := buttonSize(ResetLabel)

	headH := panelstack.LineHeight(HeadFontPx)
	cellH := panelstack.LineHeight(ValFontPx) + 2*CellPadY

	widths := make([]float32, len(rows))
	var tableW float32
	for i := range rows {
		widths[i] = colWidth(labels[i])
		tableW += widths[i]
		if i > 0 {
			tableW += ColGap
		}
	}

	contentW := startW + ButtonGap + resetW
	if tableW > contentW {
		contentW = tableW
	}
	contentH := btnH + SectionGap + headH + RowGap + cellH + RowGap + cellH

	box, x, y := st.Add(contentW, contentH)

	lay := Layout{
		Box:     box,
		Start:   Rect{X: x, Y: y, W: startW, H: btnH},
		Reset:   Rect{X: x + startW + ButtonGap, Y: y, W: resetW, H: btnH},
		Columns: make([]Column, len(rows)),
	}

	headY := y + btnH + SectionGap
	roundsY := headY + headH + RowGap
	msgsY := roundsY + cellH + RowGap

	cx := x
	for i, row := range rows {
		w := widths[i]
		lay.Columns[i] = Column{
			NodeRow: row,
			Label:   labels[i],
			Head:    Rect{X: cx, Y: headY, W: w, H: headH},
			Rounds:  Rect{X: cx, Y: roundsY, W: w, H: cellH},
			Msgs:    Rect{X: cx, Y: msgsY, W: w, H: cellH},
		}
		cx += w + ColGap
	}
	return lay
}

type Button int

const (
	ButtonNone Button = iota
	ButtonStart
	ButtonReset
)

func (l Layout) Hit(x, y float64) Button {
	if panelstack.HitRect(l.Start, x, y) {
		return ButtonStart
	}
	if panelstack.HitRect(l.Reset, x, y) {
		return ButtonReset
	}
	return ButtonNone
}
