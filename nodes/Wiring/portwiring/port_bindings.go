package portwiring

import (
	"io"

	"github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
	"github.com/dtauraso/wirefold/nodes/spatial"
	wire "github.com/dtauraso/wirefold/nodes/wire"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
	"github.com/dtauraso/wirefold/nodes/wire/outport"
)

const DriveSlotsPerNode = 2

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

func FirstPortOfDir(ports []PortSpec, dir PortDir) (string, bool) {
	for _, p := range ports {
		if p.Dir == dir {
			return p.Name, true
		}
	}
	return "", false
}

type PortBindings struct {
	singlePaced    map[string]singleBinding
	broadcastPaced map[string][]broadcastBinding

	OutSink map[string]*outport.Out

	Clock clock.Clock

	SpeedSinks *[]chan float64

	RT rowtables.RowTables

	InteriorOuts       *map[string]io.Writer
	DriveOuts          *map[string][DriveSlotsPerNode]io.Writer
	BuildInteriorFrame *func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []wire.RowEvent) []byte

	VectorOut map[string]chan tiltvector.TiltVectorMsg
	VectorIn  map[string]chan tiltvector.TiltVectorMsg
}

type singleBinding struct {
	pw    *wire.PacedWire
	rule  outport.SendRule
	steps int
	seg   spatial.WireSegment
	label string
}

type broadcastBinding struct {
	pw     *wire.PacedWire
	handle string
	rule   outport.SendRule
	steps  int
	seg    spatial.WireSegment
	label  string
}

func NewPortBindings() PortBindings {
	return PortBindings{
		singlePaced:    map[string]singleBinding{},
		broadcastPaced: map[string][]broadcastBinding{},
	}
}

func (pb *PortBindings) SetSinglePaced(name string, pw *wire.PacedWire) {
	pb.singlePaced[name] = singleBinding{pw: pw}
}

func (pb *PortBindings) SetSinglePacedRule(name string, pw *wire.PacedWire, rule outport.SendRule, steps int, seg spatial.WireSegment, label string) {
	pb.singlePaced[name] = singleBinding{pw: pw, rule: rule, steps: steps, seg: seg, label: label}
}

func (pb *PortBindings) AppendBroadcastWithHandle(name, handle string, pw *wire.PacedWire, rule outport.SendRule, steps int, seg spatial.WireSegment, label string) {
	pb.broadcastPaced[name] = append(pb.broadcastPaced[name], broadcastBinding{
		pw: pw, handle: handle, rule: rule, steps: steps, seg: seg, label: label,
	})
}

func (pb *PortBindings) deadEndIn(name string) <-chan int {
	return make(chan int, 1)
}

func (pb *PortBindings) deadEndOut(name string) chan<- int {
	return make(chan int, 1)
}

func (pb *PortBindings) deadEndOutSlice(name string) []chan<- int {
	return nil
}
