package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
)

type Net struct {
	Back  Act
	Nodes []NetNode
	IDs   []string
	Cells [][]NetCell
}

type NetNode struct {
	Name, Sub string
	Menu      Act
	MenuOpen  bool
	Drop      Board.Chip
	C         string
	Ref       []NetField
}

type NetField struct {
	Name  string
	Field Board.Field
}

type NetCell struct {
	Self, Linked bool
	Act          Act
}

const netHeadW = 104

func netBar(l *Board.Laid, r Rect, n Net) {
	clip := l.Clip(r)
	h := Board.RowH()
	l.Line(Rect{X: r.X, Y: r.Y + r.H - 0.5, W: r.W, H: 0}, clip, Board.Guide, 1, false)
	x, y := r.X+pad, r.Y+pad
	x += l.Chip(x, y, clip, Board.Chip{Label: "← steps", Act: n.Back, Px: Board.BtnPx}) + Board.Gap
	x += l.Text(x, y, h, clip, "›", Board.Px, Board.Mute, false, Board.Left) + Board.Gap
	x += l.Text(x, y, h, clip, "Network", Board.Px, Board.Ink, true, Board.Left) + Board.Gap
	if note := "shared by all sequences"; x+Board.W(note, Board.SmallPx) < r.X+r.W-pad {
		l.Text(x, y, h, clip, note, Board.SmallPx, Board.Mute, false, Board.Left)
	}
}

func network(l *Board.Laid, r Rect, sy float32, n Net, menus *[]placed) float32 {
	clip := l.Clip(r)
	l.Add(Board.Mark{Kind: Board.KBox, R: r, Clip: clip, Fill: Board.Back, Alpha: 1})
	h := Board.RowH()
	w := max(160, r.W-2*colGap)
	x, y := r.X+colGap, r.Y+colGap-sy
	card := func(name, sub string, rows int, menu Act, open bool, drop Board.Chip) (float32, float32) {
		ch := cheadH + float32(rows)*(h+Board.Gap) + cpad
		l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: x, Y: y, W: w, H: ch}, Clip: clip, Fill: Board.White, Edge: Board.Guide, Radius: radius6, Stroke: 1, Alpha: 1})
		l.Add(Board.Mark{Kind: Board.KBox, R: Rect{X: x + 1, Y: y + 1, W: w - 2, H: cheadH - 1}, Clip: clip, Fill: Board.Card, Radius: radius6 - 1, Corner: 2, Alpha: 1})
		tx := x + cpad
		tx += l.Text(tx, y+(cheadH-h)/2, h, clip, name, Board.Px, Board.Ink, true, Board.Left) + Board.Gap
		l.Text(tx, y+(cheadH-h)/2, h, clip, sub, Board.SmallPx, Board.Mute, false, Board.Left)
		if menu.Op != "" {
			at := dots(l, x+w-cpad, y, cheadH, clip, open, menu)
			if open {
				*menus = append(*menus, placed{at: at, m: Menu{Drop: drop}})
			}
		}
		return x + cpad, y + cheadH + Board.Gap
	}
	fields := func(fx, fy float32, label string, fs []NetField, tail string) {
		l.Text(fx, fy, h, clip, label, Board.SmallPx, Board.Mute, false, Board.Left)
		fx += netHeadW
		for _, f := range fs {
			fx += l.Text(fx, fy, h, clip, f.Name, Board.SmallPx, Board.Mute, false, Board.Left) + 3
			fx += l.Field(fx, fy, clip, f.Field) + Board.Gap
		}
		if tail != "" {
			l.Text(fx, fy, h, clip, tail, Board.SmallPx, Board.Mute, false, Board.Left)
		}
	}
	for _, nd := range n.Nodes {
		cx, cy := card(nd.Name, nd.Sub, 2, nd.Menu, nd.MenuOpen, nd.Drop)
		fields(cx, cy, "placed at (I)", nil, nd.C)
		fields(cx, cy+h+Board.Gap, "reference turn", nd.Ref, "")
		y = cy + 2*(h+Board.Gap) + cpad - Board.Gap + colGap
	}
	if len(n.IDs) > 1 {
		cx, cy := card("links", "each node, then the nodes it feeds · click one to unlink", len(n.IDs), Act{}, false, Board.Chip{})
		for i, row := range n.Cells {
			ry := cy + float32(i)*(h+Board.Gap)
			x := cx + l.Text(cx, ry, h, clip, n.IDs[i]+"  →", Board.Px, Board.Ink, true, Board.Left) + Board.Gap
			for j, c := range row {
				if c.Linked {
					x += l.Chip(x, ry, clip, Board.Chip{Label: n.IDs[j], Act: c.Act, Lit: true, Px: Board.BtnPx}) + Board.Gap
				}
			}
			x += Board.Gap
			for j, c := range row {
				if !c.Linked && !c.Self {
					cw := l.Chip(x, ry, clip, Board.Chip{Label: "＋" + n.IDs[j], Act: c.Act, Px: Board.BtnPx})
					l.Marks[len(l.Marks)-1].Alpha = 0.55
					x += cw + Board.Gap
				}
			}
		}
		y = cy + float32(len(n.IDs))*(h+Board.Gap) + cpad - Board.Gap + colGap
	}
	return max(0, y+sy-(r.Y+r.H))
}
