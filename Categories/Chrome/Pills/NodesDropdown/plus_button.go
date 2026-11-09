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
}

func BuildPlus(viewW float32, strip Rect, kinds []Kind) Layout {
	if len(kinds) == 0 || viewW <= 0 {
		return Layout{}
	}
	y, size := strip.Y, strip.H
	if size <= 0 {
		y, size = Panel.OriginY, plusFallbackSize
	}
	k := kinds[0]
	a, _ := NodeBuf.AppearanceOf(k.Name)
	return Layout{Plus: Plus{
		Rect:   Rect{X: Panel.OriginX, Y: y, W: size, H: size},
		KindID: k.KindID,
		Fill:   a.Fill,
		Stroke: a.Stroke,
	}}
}
