package Sheet

import "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"

const CardPad = 6

type CardBox struct {
	Box  int
	Rect Rect
	Pop  bool
	Cell int
	Max  float32
	Band Band
}

type spot struct{ X, Y float32 }

func outerW(g Group, room float32) float32 {
	if g.Card {
		return indentOf(g) + groupW(g) + 2*CardPad
	}
	if g.Band != BandNone {
		return max(room, shownW(g, room)+2*BandPad)
	}
	return shownW(g, room)
}

func outerH(g Group) float32 {
	if g.Card {
		return groupH(g) + 2*CardPad
	}
	return groupH(g)
}

func popSpot(groups []Group, spots []spot, a Anchor) spot {
	at := spots[a.Group]
	g := groups[a.Group]
	x := at.X + indentOf(g)
	if g.Band != BandNone {
		x += BandPad
	}
	for i, w := range columns(g) {
		if i == a.Col {
			break
		}
		x += w + CellGapX
	}
	return spot{x - CardPad, at.Y + groupH(g)}
}

func flow(groups []Group, room float32) (spots []spot, w, h float32) {
	var x, y, lineH, popH float32
	for i, g := range groups {
		gw, gh := outerW(g, room), outerH(g)
		if g.Pop != nil {
			s := popSpot(groups, spots, *g.Pop)
			spots = append(spots, s)
			popH = max(popH, s.Y+gh)
			continue
		}
		prev := i > 0 && groups[i-1].Card
		beside := g.Card && prev && x+Panel.Gap+gw-indentOf(g) <= room
		switch {
		case beside:
			x += Panel.Gap
			spots = append(spots, spot{x - indentOf(g), y})
			x += gw - indentOf(g)
			lineH = max(lineH, gh)
		default:
			if i > 0 {
				y += lineH + Panel.Gap
			}
			spots = append(spots, spot{0, y})
			x, lineH = gw, gh
		}
		w = max(w, x)
	}
	return spots, w, max(y+lineH, popH)
}
