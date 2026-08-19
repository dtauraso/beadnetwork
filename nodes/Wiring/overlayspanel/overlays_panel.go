package overlayspanel

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/panelstack"
	"github.com/dtauraso/wirefold/tools/topology-vscode/OverlaysDropdown"
)

type Rect = panelstack.Rect

const (
	IndentHeading = 10
	IndentRow     = 14

	CheckSize = 13
	IconW     = 11
	RowGap    = 7
)

type RowKind int

const (
	RowHeading RowKind = iota
	RowFlag
)

type Row struct {
	Kind  RowKind
	Depth int

	Rect Rect

	Heading  string
	Panel    string
	Open     bool
	CountOn  int
	CountAll int
	Count    Rect

	Flag  string
	Icon  string
	Label string
	On    bool

	Disabled bool
}

type Layout struct {
	Pill    Rect
	Open    bool
	Active  bool
	Popover Rect

	Rows []Row
}

func countOn(g Group, ov *OverlaysDropdown.OverlayState) (on, all int) {
	flags := allFlags(g)
	for _, f := range flags {
		if read, ok := OverlaysDropdown.OverlayFlagRead[f]; ok && read(ov) {
			on++
		}
	}
	return on, len(flags)
}

func Build(
	st *panelstack.PillStack,
	ov *OverlaysDropdown.OverlayState,
	pn *OverlaysDropdown.PanelState,
) Layout {
	pill := st.AddPill()
	active := OverlaysDropdown.OverlayFlagRead[GuidelinesFlag](ov)
	open := OverlaysDropdown.PanelOpen["overlays"](pn)

	lay := Layout{Pill: pill, Open: open, Active: active}
	if !open {
		st.EndGroup()
		return lay
	}

	var rows []Row
	var h float32
	var walk func(g Group, depth int)
	walk = func(g Group, depth int) {
		on, all := countOn(g, ov)
		gopen := OverlaysDropdown.PanelOpen[g.Panel](pn)
		rows = append(rows, Row{
			Kind: RowHeading, Depth: depth, Rect: Rect{Y: h, H: panelstack.HeadingH()},
			Heading: g.Heading, Panel: g.Panel, Open: gopen,
			CountOn: on, CountAll: all, Disabled: !active,
		})
		h += panelstack.HeadingH()
		if !gopen {
			return
		}
		for _, it := range g.Items {
			icon := it.Icon
			isOn := false
			if read, ok := OverlaysDropdown.OverlayFlagRead[it.Flag]; ok {
				isOn = read(ov)
			}
			if it.Flag == "labelsGlobal" {
				icon = LabelsIcon(isOn)
			}
			rows = append(rows, Row{
				Kind: RowFlag, Depth: depth, Rect: Rect{Y: h, H: panelstack.RowH()},
				Flag: it.Flag, Icon: icon, Label: it.Label, On: isOn, Disabled: !active,
			})
			h += panelstack.RowH()
		}
		for _, sub := range g.Groups {
			walk(sub, depth+1)
		}
	}
	for _, g := range Tree {
		walk(g, 0)
	}

	box, x, y := st.AddPopover(h)
	lay.Popover = box
	w := box.W - 2*panelstack.PopoverPad

	for i := range rows {
		r := &rows[i]
		r.Rect.X = x
		r.Rect.Y += y
		r.Rect.W = w
		if r.Kind == RowHeading {
			cw := panelstack.TextWidth(fmt.Sprintf("%d/%d", r.CountOn, r.CountAll), panelstack.PillHeadingPx) + 10
			r.Count = Rect{X: x + w - cw, Y: r.Rect.Y, W: cw, H: r.Rect.H}
		}
	}
	lay.Rows = rows
	st.EndGroup()
	return lay
}

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
