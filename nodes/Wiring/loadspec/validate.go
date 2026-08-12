package loadspec

import (
	"fmt"

	"strings"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/portwiring"
	"github.com/dtauraso/wirefold/nodes/wire/outport"
)

func ValidateSpec(spec *TopoSpec, kindPorts map[string][]portwiring.PortSpec) error {
	var errs []string

	kindInPorts := map[string]map[string]bool{}
	kindOutPorts := map[string]map[string]bool{}
	kindBroadcastPorts := map[string]map[string]bool{}
	for kind, ports := range kindPorts {
		ins := map[string]bool{}
		outs := map[string]bool{}
		outMultis := map[string]bool{}
		for _, p := range ports {
			switch p.Dir {
			case portwiring.PortIn:
				ins[p.Name] = true
			case portwiring.PortOut:
				outs[p.Name] = true
			case portwiring.PortBroadcast:
				outMultis[p.Name] = true
				outs[p.Name] = true
			}
		}
		kindInPorts[kind] = ins
		kindOutPorts[kind] = outs
		kindBroadcastPorts[kind] = outMultis
	}

	nodeType := map[string]string{}
	seenID := map[string]bool{}
	for _, n := range spec.Nodes {
		if seenID[n.ID] {

			errs = append(errs, fmt.Sprintf("duplicate node id %q", n.ID))
			continue
		}
		seenID[n.ID] = true
		nodeType[n.ID] = n.Type
		if _, ok := kindPorts[n.Type]; !ok {
			errs = append(errs, fmt.Sprintf("node %q: unknown type %q", n.ID, n.Type))
		}
	}

	for _, n := range spec.Nodes {
		if !jsonpersist.SafeTreePathComponent(n.ID) {
			errs = append(errs, fmt.Sprintf("node id %q is not a safe path component", n.ID))
		}
	}

	for _, e := range spec.Edges {
		if e.Label == "" {
			errs = append(errs, fmt.Sprintf("edge %q→%q has empty label", e.Source, e.Target))
		}
	}

	for _, e := range spec.Edges {
		srcKind, srcKnown := nodeType[e.Source]
		if !srcKnown {
			errs = append(errs, fmt.Sprintf("edge %q references unknown node id %q as its source", e.Label, e.Source))
		} else {
			srcHandle := e.SourceHandle
			if base, isMulti := BroadcastBaseName(srcHandle, srcKind, kindBroadcastPorts); isMulti {
				srcHandle = base
			}
			if !kindOutPorts[srcKind][srcHandle] {
				errs = append(errs, fmt.Sprintf("edge %q: sourceHandle %q is not an output port on kind %q", e.Label, e.SourceHandle, srcKind))
			}
		}
		tgtKind, tgtKnown := nodeType[e.Target]
		if !tgtKnown {
			errs = append(errs, fmt.Sprintf("edge %q references unknown node id %q as its target", e.Label, e.Target))
		} else if !kindInPorts[tgtKind][e.TargetHandle] {
			errs = append(errs, fmt.Sprintf("edge %q: targetHandle %q is not an input port on kind %q", e.Label, e.TargetHandle, tgtKind))
		}
	}

	for _, n := range spec.Nodes {
		if n.Data == nil || n.Data.SendRules == nil {
			continue
		}
		for port, raw := range n.Data.SendRules {
			if _, err := outport.ParseSendRule(raw); err != nil {
				errs = append(errs, fmt.Sprintf("node %q port %q: %v", n.ID, port, err))
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("LoadTopology: spec validation failed:\n  %s", strings.Join(errs, "\n  "))
}
