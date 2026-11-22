package LinksBoard

import (
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const (
	treeW      = 200
	treeIndent = 12
	foldW      = 14
)

type TreeRow struct {
	Label    string
	Depth    int
	Kids     int
	Add      bool
	Open     bool
	Here     bool
	Shown    bool
	Go       Act
	Fold     Act
	Menu     Act
	Name     Act
	MenuOpen bool
	Items    []Board.Chip
	Drop     Board.Chip
}

func treePane(l *Board.Laid, r Rect, sy float32, d Data, menus *[]placed) float32 {
	clip := l.Clip(r)
	l.Add(Board.Mark{Kind: Board.KBox, R: r, Clip: clip, Fill: Board.Card, Alpha: 1})
	l.Line(Rect{X: r.X + r.W - 0.5, Y: r.Y, W: 0, H: r.H}, clip, Board.Guide, 1, false)
	h := Board.RowH()
	y := r.Y + pad - sy
	if d.TreeFind != nil {
		l.Field(r.X+pad, y, clip, *d.TreeFind)
		y += h + 4
	}
	for _, t := range d.Tree {
		row := Rect{X: r.X + 2, Y: y, W: r.W - 4, H: h}
		if t.Shown {
			l.Add(Board.Mark{Kind: Board.KBox, R: row, Clip: clip, Fill: Board.On, Radius: 4, Alpha: 1})
		}
		x := r.X + pad + float32(t.Depth)*treeIndent
		if t.Kids > 0 {
			g := "›"
			if t.Open {
				g = "⌄"
			}
			l.Text(x+foldW/2, y, h, clip, g, Board.Px, Board.Mute, false, Board.Center)
			l.Spot(Rect{X: x, Y: y, W: foldW, H: h}, clip, Sheet.Cell{Act: t.Fold}, Sheet.PartBody)
		}
		x += foldW
		right := r.X + r.W - pad
		if t.Add {
			l.Text(x, y, h, clip, "＋ add step", 12, Board.Mute, false, Board.Left)
			l.Spot(Rect{X: x, Y: y, W: max(0, right-x), H: h}, clip, Sheet.Cell{Act: t.Go}, Sheet.PartBody)
			y += h
			continue
		}
		if t.Menu.Op != "" {
			at := dots(l, right, y, h, clip, t.MenuOpen, t.Menu)
			right = at.X - 2
			if t.MenuOpen {
				*menus = append(*menus, placed{at: at, m: Menu{Items: t.Items, Drop: t.Drop}})
			}
		}
		if t.Kids > 0 {
			n := strconv.Itoa(t.Kids)
			l.Text(right-2, y, h, clip, n, 11, Board.Mute, false, Board.Right)
			right -= Board.W(n, 11) + 8
		}
		ink := Board.Ink
		if !t.Here {
			ink = Board.Mute
		}
		l.Text(x, y, h, clip, fitText(t.Label, right-x-4, t.Here), Board.Px, ink, t.Here, Board.Left)
		l.Marks[len(l.Marks)-1].Edit = t.Name
		l.Spot(Rect{X: x, Y: y, W: max(0, right-x), H: h}, clip, Sheet.Cell{Act: t.Go}, Sheet.PartBody)
		y += h
	}
	if len(d.Tree) == 0 && d.TreeEmpty != "" {
		l.Text(r.X+pad, y, h, clip, d.TreeEmpty, 12, Board.Mute, false, Board.Left)
		y += h
	}
	return max(0, y+pad+sy-(r.Y+r.H))
}

func treeMost(r Rect, d Data) float32 {
	h := Board.RowH()
	y := 2*pad + float32(len(d.Tree))*h
	if d.TreeFind != nil {
		y += h + 4
	}
	if len(d.Tree) == 0 && d.TreeEmpty != "" {
		y += h
	}
	return max(0, y-r.H)
}

func fitText(s string, w float32, bold bool) string {
	wt := map[bool]uint8{true: 7, false: 4}[bold]
	if Board.TextW(s, Board.Px, wt) <= w {
		return s
	}
	rs := []rune(s)
	for len(rs) > 0 && Board.TextW(string(rs)+"…", Board.Px, wt) > w {
		rs = rs[:len(rs)-1]
	}
	return string(rs) + "…"
}
