package moverreg

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/kindreg"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
)

func (mr *MoverRegistry) LinkRefusal(src, kind string) (srcPort, targetPort, why string, ok bool) {
	srcGeom, found := mr.nodeGeoms[src]
	srcKind := ""
	if found {
		srcKind = srcGeom.Kind()
	}
	return linkRefusalFor(src, srcKind, found, kind)
}

func linkRefusalFor(src, srcKind string, srcFound bool, kind string) (srcPort, targetPort, why string, ok bool) {
	targetPort, hasIn := firstPortOfDir(kind, portwiring.PortIn)
	if !hasIn {
		return "", "", fmt.Sprintf("%s takes no input, so nothing can connect to it", kind), false
	}
	if !srcFound {
		return "", "", fmt.Sprintf("no geometry for %s", src), false
	}
	srcPort, hasOut := firstPortOfDir(srcKind, portwiring.PortOut)
	if !hasOut {
		return "", "", fmt.Sprintf("%s has no output to connect from", srcKind), false
	}
	return srcPort, targetPort, "", true
}

func firstPortOfDir(kind string, dir portwiring.PortDir) (string, bool) {
	b, ok := kindreg.Registry[kind]
	if !ok {
		return "", false
	}
	return portwiring.FirstPortOfDir(b.Ports, dir)
}
