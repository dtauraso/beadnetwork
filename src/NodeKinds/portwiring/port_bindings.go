package portwiring

import (
	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/src/Clock"
	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/src/Node/Interior"
	"github.com/dtauraso/wirefold/src/Scene/rowtables"
	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
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

	OutSink map[string]*beadanimation.Sender

	Clock clock.Clock

	SpeedSinks *SliderPanel.Sinks

	RT rowtables.RowTables

	InteriorEmitters   *map[string]*interior.Emitter
	BuildInteriorFrame *func(tick uint32, events []B.RowEvent) []byte

	VectorOut map[string]chan TiltPanel.TiltVectorMsg
	VectorIn  map[string]chan TiltPanel.TiltVectorMsg
}

type singleBinding struct {
	pw    *beadanimation.BeadLine
	rule  beadanimation.SendRule
	label string
}

type broadcastBinding struct {
	pw     *beadanimation.BeadLine
	handle string
	rule   beadanimation.SendRule
	label  string
}

func NewPortBindings() PortBindings {
	return PortBindings{
		singlePaced:    map[string]singleBinding{},
		broadcastPaced: map[string][]broadcastBinding{},
	}
}

func (pb *PortBindings) SetSinglePaced(name string, pw *beadanimation.BeadLine) {
	pb.singlePaced[name] = singleBinding{pw: pw}
}

func (pb *PortBindings) SetSinglePacedRule(name string, pw *beadanimation.BeadLine, rule beadanimation.SendRule, label string) {
	pb.singlePaced[name] = singleBinding{pw: pw, rule: rule, label: label}
}

func (pb *PortBindings) AppendBroadcastWithHandle(name, handle string, pw *beadanimation.BeadLine, rule beadanimation.SendRule, label string) {
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
