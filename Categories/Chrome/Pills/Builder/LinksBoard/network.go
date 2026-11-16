package LinksBoard

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
)

type Net struct {
	Back, Add Act
	Nodes     []NetNode
	IDs       []string
	Cells     [][]NetCell
}

type NetNode struct {
	Name, Sub string
	Menu      Act
	MenuOpen  bool
	Drop      Board.Chip
	Place     []NetField
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
	add := Board.Chip{Label: "＋ node", Act: n.Add, Lit: true, Px: Board.BtnPx}
	aw := Board.W(add.Label, Board.BtnPx) + 2*Board.PadX
	if right := r.X + r.W - pad - aw; right > x {
		if note := "shared by all sequences"; x+Board.W(note, Board.SmallPx) < right-Board.Gap {
			l.Text(x, y, h, clip, note, Board.SmallPx, Board.Mute, false, Board.Left)
		}
		l.Chip(right, y, clip, add)
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
		fields(cx, cy, "placed at (I)", nd.Place, nd.C)
		fields(cx, cy+h+Board.Gap, "reference turn", nd.Ref, "")
		y = cy + 2*(h+Board.Gap) + cpad - Board.Gap + colGap
	}
	if len(n.IDs) > 1 {
		cx, cy := card("links", "row from, column into", len(n.IDs)+1, Act{}, false, Board.Chip{})
		cw := float32(36)
		for j, id := range n.IDs {
			l.Text(cx+cw*float32(j+1)+cw/2, cy, h, clip, "→"+id, Board.SmallPx, Board.Mute, true, Board.Center)
		}
		for i, row := range n.Cells {
			ry := cy + float32(i+1)*(h+Board.Gap)
			l.Text(cx+cw/2, ry, h, clip, n.IDs[i], Board.SmallPx, Board.Mute, true, Board.Center)
			for j, c := range row {
				mx := cx + cw*float32(j+1)
				switch {
				case c.Self:
					l.Text(mx+cw/2, ry, h, clip, "—", Board.SmallPx, Board.Faint, false, Board.Center)
				case c.Linked:
					l.Chip(mx+4, ry, clip, Board.Chip{Label: "●", Act: c.Act, Lit: true, Px: Board.BtnPx})
				default:
					l.Chip(mx+4, ry, clip, Board.Chip{Label: "＋", Act: c.Act, Px: Board.BtnPx})
				}
			}
		}
		y = cy + float32(len(n.IDs)+1)*(h+Board.Gap) + cpad - Board.Gap + colGap
	}
	return max(0, y+sy-(r.Y+r.H))
}
