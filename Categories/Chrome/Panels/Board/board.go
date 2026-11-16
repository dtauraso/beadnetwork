package Board

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

type Rect = Panel.Rect

type Paint uint8

const (
	None Paint = iota
	Ink
	Mute
	Card
	Guide
	On
	Edge
	Accent
	Phi
	Theta
	White
	Warn
	WarnFill
	StartArc
	Hot
	Faint
	Back
	Hair
	DimInk
	Bad
)

type Kind uint8

const (
	KBox Kind = iota
	KText
	KDial
	KLine
	KCircle
)

type Align uint8

const (
	Left Align = iota
	Center
	Right
)

type Dial struct {
	Axis               Paint
	Start, Turn, Ghost float32
	Inc                float32
	Labels             bool
}

type Mark struct {
	Kind   Kind
	R      Rect
	Clip   uint8
	Fill   Paint
	Edge   Paint
	Ink    Paint
	Radius float32
	Stroke float32
	Text   string
	Px     float32
	Bold   bool
	Align  Align
	Alpha  float32
	Dash   bool
	Weight uint8
	Corner uint8
	Dial   Dial
	Edit   Sheet.Act
}

type Spot struct {
	R    Rect
	Clip uint8
	Cell Sheet.Cell
	Part Sheet.Part
}

type Laid struct {
	Marks        []Mark
	Spots        []Spot
	Clips        []Rect
	MostX, MostY float32
}

func New(box Rect) *Laid { return &Laid{Clips: []Rect{box}} }

func (l *Laid) Clip(r Rect) uint8 {
	l.Clips = append(l.Clips, r)
	return uint8(len(l.Clips) - 1)
}

func (l *Laid) Add(m Mark) { l.Marks = append(l.Marks, m) }

func (l *Laid) Spot(r Rect, clip uint8, c Sheet.Cell, p Sheet.Part) {
	if c.Act.Op == "" {
		return
	}
	l.Spots = append(l.Spots, Spot{R: r, Clip: clip, Cell: c, Part: p})
}

func (l *Laid) Block(r Rect, clip uint8) {
	l.Spots = append(l.Spots, Spot{R: r, Clip: clip, Part: Sheet.PartInside})
}

func (l *Laid) Most() (float32, float32) { return l.MostX, l.MostY }

func (l *Laid) Hit(x, y float64) (Sheet.Hit, bool) {
	for i := len(l.Spots) - 1; i >= 0; i-- {
		s := l.Spots[i]
		if int(s.Clip) < len(l.Clips) && !Panel.HitRect(l.Clips[s.Clip], x, y) {
			continue
		}
		if Panel.HitRect(s.R, x, y) {
			return Sheet.Hit{Part: s.Part, Cell: s.Cell, Rect: s.R}, true
		}
	}
	return Sheet.Hit{}, false
}

func Clamp(v, most float32) float32 { return max(0, min(v, most)) }
