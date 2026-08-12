// mover_registry_linkrefusal.go — the pure "can this new edge exist" decision
// (LinkRefusal) plus its two pure helpers. Split out of mover_registry.go by concern —
// see that file's header for the split rationale. linkRefusalFor/firstPortOfDir never
// touch a MoverRegistry field once handed src's resolved kind, so this whole file is one
// self-contained decision, not state ownership.
package moverreg

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/kindapi"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
)

// LinkRefusal answers whether an edge from src to a NEW node of kind can exist, and says
// why not when it cannot. mr's only part is resolving src's own kind off nodeGeoms; the
// two structural reasons themselves are decided by the pure linkRefusalFor below.
func (mr *MoverRegistry) LinkRefusal(src, kind string) (srcPort, targetPort, why string, ok bool) {
	srcGeom, found := mr.nodeGeoms[src]
	srcKind := ""
	if found {
		srcKind = srcGeom.Kind()
	}
	return linkRefusalFor(src, srcKind, found, kind)
}

// linkRefusalFor is the pure decision LinkRefusal makes once src's own kind (and whether
// it was found at all) has been resolved: kind must take an input, and src must have
// both geometry and an output to connect from. Split out of LinkRefusal because it never
// touched MoverRegistry itself, only the two node kinds it was handed.
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

// firstPortOfDir looks up kind's registered ports (kindapi.Registry) and forwards to
// portwiring.FirstPortOfDir (pure over a []PortSpec, no dispatch-core state) for the FIRST
// port in dir, in the order the kind declared them at RegisterBuilder. First, not "In": a
// kind names its own ports, and the declaration order is the only ranking there is —
// NormalSum's NormalA before NormalB says which one an edge should take when nothing else
// has been said. Moved here (from package dispatch's scene_structure.go) alongside
// linkRefusalFor, its only caller.
func firstPortOfDir(kind string, dir portwiring.PortDir) (string, bool) {
	b, ok := kindapi.Registry[kind]
	if !ok {
		return "", false
	}
	return portwiring.FirstPortOfDir(b.Ports, dir)
}
