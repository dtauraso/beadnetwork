package speedpanel

type Setting struct {
	Speed float64

	Num string
	Den string
}

var Settings = []Setting{
	{Speed: 0, Num: "0"},
	{Speed: 0.25, Num: "1", Den: "4"},
	{Speed: 0.5, Num: "1", Den: "2"},
	{Speed: 0.75, Num: "3", Den: "4"},
	{Speed: 1, Num: "1"},
	{Speed: 2, Num: "2"},
}

const (
	TrackW     = 104
	ThumbInset = 6

	TrackH = 4

	TickW = 14
	TickH = 12

	OriginX = 12
	OriginY = 12
)

type Rect struct{ X, Y, W, H float32 }

func Layout() (ticks []Rect, track Rect) {
	track = Rect{X: OriginX, Y: OriginY, W: TrackW, H: TrackH}

	span := float64(TrackW - 2*ThumbInset)
	ticks = make([]Rect, len(Settings))
	for i := range Settings {
		centre := float64(OriginX + ThumbInset)
		if len(Settings) > 1 {
			centre += span * float64(i) / float64(len(Settings)-1)
		}
		ticks[i] = Rect{
			X: float32(centre) - TickW/2,
			Y: OriginY + TrackH + 2,
			W: TickW,
			H: TickH,
		}
	}
	return ticks, track
}

func SelectedIndex(speed float64) int {
	best, bestDiff := 0, -1.0
	for i, s := range Settings {
		d := s.Speed - speed
		if d < 0 {
			d = -d
		}
		if bestDiff < 0 || d < bestDiff {
			best, bestDiff = i, d
		}
	}
	return best
}

func Hit(x, y float64) int {
	ticks, _ := Layout()
	for i, r := range ticks {
		if x >= float64(r.X) && x <= float64(r.X+r.W) && y >= float64(r.Y) && y <= float64(r.Y+r.H) {
			return i
		}
	}
	return -1
}
