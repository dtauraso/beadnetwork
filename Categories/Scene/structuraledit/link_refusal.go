package structuraledit

import (
	"fmt"

	"github.com/dtauraso/wirefold/Categories/NodeKinds/portwiring"
)

type Link struct {
	SrcPort    string
	TargetPort string
	Broadcast  bool
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
	ports, ok := portwiring.KindPorts[kind]
	if !ok {
		return "", false
	}
	return portwiring.FirstPortOfDir(ports, dir)
}
