package speedpanel

import "github.com/dtauraso/wirefold/nodes/Wiring/panelstack"

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

	TrackTickGap = 2
)

type Rect = panelstack.Rect

type Layout struct {
	Box   Rect
	Track Rect
	Ticks []Rect
}

func Build(st *panelstack.Stack) Layout {
	box, x, y := st.Add(TrackW, TrackH+TrackTickGap+TickH)

	lay := Layout{
		Box:   box,
		Track: Rect{X: x, Y: y, W: TrackW, H: TrackH},
		Ticks: make([]Rect, len(Settings)),
	}

	span := float64(TrackW - 2*ThumbInset)
	for i := range Settings {
		centre := float64(x + ThumbInset)
		if len(Settings) > 1 {
			centre += span * float64(i) / float64(len(Settings)-1)
		}
		lay.Ticks[i] = Rect{
			X: float32(centre) - TickW/2,
			Y: y + TrackH + TrackTickGap,
			W: TickW,
			H: TickH,
		}
	}
	return lay
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

func (l Layout) Hit(x, y float64) int {
	for i, r := range l.Ticks {
		if panelstack.HitRect(r, x, y) {
			return i
		}
	}
	return -1
}
