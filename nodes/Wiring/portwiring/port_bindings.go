package portwiring

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/interior"
	"github.com/dtauraso/wirefold/nodes/Wiring/rowtables"
	"github.com/dtauraso/wirefold/nodes/Wiring/tiltvector"
	"github.com/dtauraso/wirefold/nodes/bead"
	"github.com/dtauraso/wirefold/nodes/bead/outport"
	"github.com/dtauraso/wirefold/nodes/clock"
	"github.com/dtauraso/wirefold/nodes/rowevent"
	"github.com/dtauraso/wirefold/tools/topology-vscode/Slider"
)

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

	SpeedSinks *Slider.Sinks

	RT rowtables.RowTables

	InteriorEmitters   *map[string]*interior.Emitter
	BuildInteriorFrame *func(tick uint32, present []uint8, value []int32, ox, oy, oz []float32, events []rowevent.RowEvent) []byte

	VectorOut map[string]chan tiltvector.TiltVectorMsg
	VectorIn  map[string]chan tiltvector.TiltVectorMsg
}

type singleBinding struct {
	pw    *bead.BeadRun
	rule  outport.SendRule
	label string
}

type broadcastBinding struct {
	pw     *bead.BeadRun
	handle string
	rule   outport.SendRule
	label  string
}

func NewPortBindings() PortBindings {
	return PortBindings{
		singlePaced:    map[string]singleBinding{},
		broadcastPaced: map[string][]broadcastBinding{},
	}
}

func (pb *PortBindings) SetSinglePaced(name string, pw *bead.BeadRun) {
	pb.singlePaced[name] = singleBinding{pw: pw}
}

func (pb *PortBindings) SetSinglePacedRule(name string, pw *bead.BeadRun, rule outport.SendRule, label string) {
	pb.singlePaced[name] = singleBinding{pw: pw, rule: rule, label: label}
}

func (pb *PortBindings) AppendBroadcastWithHandle(name, handle string, pw *bead.BeadRun, rule outport.SendRule, label string) {
	pb.broadcastPaced[name] = append(pb.broadcastPaced[name], broadcastBinding{
		pw: pw, handle: handle, rule: rule, label: label,
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
