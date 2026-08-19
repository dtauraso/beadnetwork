package inputcodec

import (
	"github.com/dtauraso/wirefold/nodes/bead"
)

type EdgeEndpoints struct {
	Source       string
	Target       string
	SourceHandle string
	TargetHandle string
}

type StdinMsg struct {
	Type string
	Op   string
	Kind string
	Attr string
	Flag string

	Num int

	X, Y float64

	Event *RawInputMsg
}

type RawInputMsg struct {
	Kind       string
	X          float64
	Y          float64
	RectLeft   float64
	RectTop    float64
	RectWidth  float64
	RectHeight float64
	Button     int
	Ctrl       bool
	Shift      bool
	Alt        bool
	Meta       bool
	DeltaX     float64
	DeltaY     float64

	Hit RawHit

	Key string
}

type RawHit struct {
	Kind string

	PortRow int

	EdgeRow int

	NodeRow int
	IsInput bool
}

type SlotRegistry map[string]*bead.BeadRun
