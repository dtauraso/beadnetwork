package NodesDropdown

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	NodeBuf "github.com/dtauraso/beadnetwork/Categories/Node"
)

const plusFallbackSize = 24

type Plus struct {
	Rect   Rect
	KindID uint8
	Fill   string
	Stroke string

	Menu bool
}

func BuildPlus(viewW float32, strip Rect, open bool, kinds []Kind) Layout {
	if len(kinds) == 0 || viewW <= 0 {
		return Layout{}
	}
	y, size := strip.Y, strip.H
	if size <= 0 {
		y, size = Panel.OriginY, plusFallbackSize
	}
	k := kinds[0]
	a, _ := NodeBuf.AppearanceOf(k.Name)
	plus := Plus{
		Rect:   Rect{X: Panel.OriginX, Y: y, W: size, H: size},
		KindID: k.KindID,
		Fill:   a.Fill,
		Stroke: a.Stroke,
		Menu:   len(kinds) > 1,
	}
	lay := Layout{Plus: plus}
	if !plus.Menu || !open {
		return lay
	}
	w := menuWidth(kinds)
	lay.Open = true
	lay.Popover = Rect{X: plus.Rect.X, Y: plus.Rect.Y + plus.Rect.H + Panel.PopoverGap, W: w + 2*Panel.PopoverPad}
	lay.Popover.H = rowsHeight(kinds, w) + 2*Panel.PopoverPad
	lay.Rows = layRows(kinds, lay.Popover.X+Panel.PopoverPad, lay.Popover.Y+Panel.PopoverPad, w)
	return lay
}

func menuWidth(kinds []Kind) float32 {
	var t float32
	for _, k := range kinds {
		t = max(t, Panel.TextWidth(k.Name, Panel.PillFontPx))
	}
	return 2*Panel.RowPadX + GlyphW + SwatchGap + SwatchSize + SwatchGap + t
}
