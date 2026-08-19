package overlayspanel

import "github.com/dtauraso/wirefold/src/Node/Wiring/panelstack"

type HitKind int

const (
	HitNone HitKind = iota
	HitPillBody
	HitPillCaret
	HitHeading
	HitCount
	HitFlag
)

type Hit struct {
	Kind  HitKind
	Panel string
	Flag  string

	Flags  []string
	Target bool

	Rect     Rect
	Tip      string
	Disabled bool
}

func (l Layout) Hit(x, y float64) Hit {
	if panelstack.HitRect(l.Pill, x, y) {
		caret := Rect{X: l.Pill.X + l.Pill.W - panelstack.CaretW, Y: l.Pill.Y, W: panelstack.CaretW, H: l.Pill.H}
		if panelstack.HitRect(caret, x, y) {
			tip := "Open overlay list"
			if l.Open {
				tip = "Close overlay list"
			}
			return Hit{Kind: HitPillCaret, Panel: "overlays", Rect: caret, Tip: tip}
		}
		return Hit{Kind: HitPillBody, Flag: GuidelinesFlag, Rect: l.Pill, Tip: "Toggle guidelines"}
	}
	if !l.Open {
		return Hit{}
	}
	if !panelstack.HitRect(l.Popover, x, y) {
		return Hit{}
	}
	for _, r := range l.Rows {
		if !panelstack.HitRect(r.Rect, x, y) {
			continue
		}
		if r.Kind == RowHeading {
			if !r.Disabled && panelstack.HitRect(r.Count, x, y) {
				tip := "Turn all " + r.Heading + " off"
				if r.CountOn == 0 {
					tip = "Turn all " + r.Heading + " on"
				}
				return Hit{
					Kind: HitCount, Flags: headingFlags(r.Heading), Target: r.CountOn == 0,
					Rect: r.Count, Tip: tip,
				}
			}
			tip := "Expand " + r.Heading
			if r.Open {
				tip = "Collapse " + r.Heading
			}
			return Hit{Kind: HitHeading, Panel: r.Panel, Rect: r.Rect, Tip: tip}
		}
		if r.Disabled {
			return Hit{Rect: r.Rect, Disabled: true}
		}
		return Hit{Kind: HitFlag, Flag: r.Flag, Rect: r.Rect, Tip: r.Label}
	}
	return Hit{}
}

func headingFlags(heading string) []string {
	var find func(gs []Group) []string
	find = func(gs []Group) []string {
		for _, g := range gs {
			if g.Heading == heading {
				return allFlags(g)
			}
			if got := find(g.Groups); got != nil {
				return got
			}
		}
		return nil
	}
	return find(Tree)
}
