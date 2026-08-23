package Drag

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

func DecodeRawInput(rec []byte) (RawInputMsg, bool) {
	if len(rec) == 0 || rec[0] != KindRawInput {
		return RawInputMsg{}, false
	}
	return decodeRawInputFrom(NewReader(rec, 1))
}
