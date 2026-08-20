package Pills

import (
	"fmt"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/src/Overlay"
)

type Rect = Panel.Rect

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

	Scroll    float32
	MaxScroll float32

	Rows []Row
}

func countOn(g Group, ov *Overlay.OverlayState) (on, all int) {
	flags := allFlags(g)
	for _, f := range flags {
		if read, ok := Overlay.OverlayFlagRead[f]; ok && read(ov) {
			on++
		}
	}
	return on, len(flags)
}

func Build(
	st *Panel.PillStack,
	ov *Overlay.OverlayState,
	pn *Panel.PanelState,
	scroll float32,
) Layout {
	pill := st.AddPill()
	active := Overlay.OverlayFlagRead[GuidelinesFlag](ov)
	open := Panel.PanelOpen["overlays"](pn)

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
		gopen := Panel.PanelOpen[g.Panel](pn)
		rows = append(rows, Row{
			Kind: RowHeading, Depth: depth, Rect: Rect{Y: h, H: Panel.HeadingH()},
			Heading: g.Heading, Panel: g.Panel, Open: gopen,
			CountOn: on, CountAll: all, Disabled: !active,
		})
		h += Panel.HeadingH()
		if !gopen {
			return
		}
		for _, it := range g.Items {
			icon := it.Icon
			isOn := false
			if read, ok := Overlay.OverlayFlagRead[it.Flag]; ok {
				isOn = read(ov)
			}
			if it.Flag == "labelsGlobal" {
				icon = LabelsIcon(isOn)
			}
			rows = append(rows, Row{
				Kind: RowFlag, Depth: depth, Rect: Rect{Y: h, H: Panel.RowH()},
				Flag: it.Flag, Icon: icon, Label: it.Label, On: isOn, Disabled: !active,
			})
			h += Panel.RowH()
		}
		for _, sub := range g.Groups {
			walk(sub, depth+1)
		}
	}
	for _, g := range Tree {
		walk(g, 0)
	}

	box, x, y, maxScroll := st.AddScrollingPopover(h, scroll)
	lay.Popover = box
	lay.MaxScroll = maxScroll
	if scroll > maxScroll {
		scroll = maxScroll
	}
	lay.Scroll = scroll
	w := box.W - 2*Panel.PopoverPad

	for i := range rows {
		r := &rows[i]
		r.Rect.X = x
		r.Rect.Y += y
		r.Rect.W = w
		if r.Kind == RowHeading {
			cw := Panel.TextWidth(fmt.Sprintf("%d/%d", r.CountOn, r.CountAll), Panel.PillHeadingPx) + 10
			r.Count = Rect{X: x + w - cw, Y: r.Rect.Y, W: cw, H: r.Rect.H}
		}
	}
	lay.Rows = rows
	st.EndGroup()
	return lay
}
