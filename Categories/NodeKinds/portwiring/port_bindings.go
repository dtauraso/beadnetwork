package portwiring

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/TiltPanel"
	clock "github.com/dtauraso/wirefold/Categories/Clock"
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
	interior "github.com/dtauraso/wirefold/Categories/Node/Interior"
	"github.com/dtauraso/wirefold/Categories/Scene/rowtables"
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

	InteriorEmitters *map[string]*interior.Emitter

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

func (pb PortBindings) SinglePacedOf(name string) (*beadanimation.BeadLine, beadanimation.SendRule, string) {
	b := pb.singlePaced[name]
	return b.pw, b.rule, b.label
}

func (pb PortBindings) BroadcastCountOf(name string) int { return len(pb.broadcastPaced[name]) }

func (pb PortBindings) BroadcastAt(name string, i int) (*beadanimation.BeadLine, string, beadanimation.SendRule, string) {
	b := pb.broadcastPaced[name][i]
	return b.pw, b.handle, b.rule, b.label
}

func (pb PortBindings) NodeRowFor(id string) (int32, bool) { return pb.RT.NodeRowFor(id) }

func (pb PortBindings) SetOutSink(key string, o *beadanimation.Sender) {
	if pb.OutSink != nil {
		pb.OutSink[key] = o
	}
}

func (pb PortBindings) InteriorEmitterOf(name string) *interior.Emitter {
	if pb.InteriorEmitters == nil || *pb.InteriorEmitters == nil {
		return nil
	}
	return (*pb.InteriorEmitters)[name]
}

func (pb PortBindings) ClockOf() clock.Clock { return pb.Clock }

func (pb PortBindings) SpeedSinksOf() *SliderPanel.Sinks { return pb.SpeedSinks }

func (pb PortBindings) VectorOutOf(name string) chan<- TiltPanel.TiltVectorMsg {
	if pb.VectorOut == nil {
		return nil
	}
	return pb.VectorOut[name]
}

func (pb PortBindings) VectorInOf(name string) <-chan TiltPanel.TiltVectorMsg {
	if pb.VectorIn == nil {
		return nil
	}
	return pb.VectorIn[name]
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
