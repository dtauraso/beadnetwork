package Panel

const (
	OriginX = 8
	OriginY = 8

	Gap = 6

	PadX = 10
	PadY = 6

	Radius = 6
)

type Rect struct{ X, Y, W, H float32 }

type Stack struct {
	y     float32
	viewH float32
}

func New(viewH float32) *Stack { return &Stack{y: OriginY, viewH: viewH} }

func (s *Stack) RoomBelow(y float32) float32 { return RoomBelow(s.viewH, y, Gap) }

func (s *Stack) Next() float32 { return s.y }

func (s *Stack) Took(h float32) { s.y += h + Gap }

func (s *Stack) Add(contentW, contentH float32) (box Rect, contentX, contentY float32) {
	box = Rect{
		X: OriginX,
		Y: s.y,
		W: contentW + 2*PadX,
		H: contentH + 2*PadY,
	}
	s.y += box.H + Gap
	return box, box.X + PadX, box.Y + PadY
}

func Advance(fontPx float32) float32 { return fontPx * 0.55 }

func TextWidth(s string, fontPx float32) float32 { return Advance(fontPx) * float32(len(s)) }

func LineHeight(fontPx float32) float32 { return fontPx * 1.2 }

func HitRect(r Rect, x, y float64) bool {
	return x >= float64(r.X) && x <= float64(r.X+r.W) && y >= float64(r.Y) && y <= float64(r.Y+r.H)
}
