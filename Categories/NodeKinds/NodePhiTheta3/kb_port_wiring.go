package NodePhiTheta3

import (
	SliderPanel "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
)

type bindings interface {
	ClockOf() clock.Clock
	SpeedSinksOf() *SliderPanel.Sinks
	VectorOutOf(name string) chan<- TiltPanel.TiltVectorMsg
	VectorInOf(name string) <-chan TiltPanel.TiltVectorMsg
}

type PortDir int

const (
	PortIn PortDir = iota
	PortOut
	PortBroadcast
)

type PortSpec struct {
	Name string
	Dir  PortDir
}
