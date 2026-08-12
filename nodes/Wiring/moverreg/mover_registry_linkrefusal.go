package moverreg

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/kindreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
)

// Link is the port pair a dropped node connects through. Broadcast says which of the two
// output directions SrcPort is: a PortOut carries one edge and a PortBroadcast carries many,
// and the handle written to disk is indexed for the second kind only.
type Link struct {
	SrcPort    string
	TargetPort string
	Broadcast  bool
}

func (mr *MoverRegistry) LinkRefusal(src, kind string) (Link, string, bool) {
	srcGeom, found := mr.nodeGeoms[src]
	srcKind := ""
	if found {
		srcKind = srcGeom.Kind()
	}
	return linkRefusalFor(src, srcKind, found, kind)
}

func linkRefusalFor(src, srcKind string, srcFound bool, kind string) (Link, string, bool) {
	targetPort, hasIn := firstPortOfDir(kind, portwiring.PortIn)
	if !hasIn {
		return Link{}, fmt.Sprintf("%s takes no input, so nothing can connect to it", kind), false
	}
	if !srcFound {
		return Link{}, fmt.Sprintf("no geometry for %s", src), false
	}
	srcPort, broadcast, hasOut := firstOutputPort(srcKind)
	if !hasOut {
		return Link{}, fmt.Sprintf("%s has no output to connect from", srcKind), false
	}
	return Link{SrcPort: srcPort, TargetPort: targetPort, Broadcast: broadcast}, "", true
}

// firstOutputPort answers "what does this kind send from", which PortOut alone does not:
// build_nodes.go binds PortBroadcast as an outbound port too, so a kind whose only output is
// broadcast (Time, TimeStart) has one. Asking for PortOut alone told the ring scene that
// Time had no output to connect from and refused every node dropped beside it, while the
// pair scene never noticed because PairNode's single output is a PortOut.
func firstOutputPort(kind string) (name string, broadcast, ok bool) {
	if name, ok = firstPortOfDir(kind, portwiring.PortOut); ok {
		return name, false, true
	}
	if name, ok = firstPortOfDir(kind, portwiring.PortBroadcast); ok {
		return name, true, true
	}
	return "", false, false
}

func firstPortOfDir(kind string, dir portwiring.PortDir) (string, bool) {
	b, ok := kindreg.Registry[kind]
	if !ok {
		return "", false
	}
	return portwiring.FirstPortOfDir(b.Ports, dir)
}
