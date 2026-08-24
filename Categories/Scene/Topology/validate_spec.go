package Topology

import (
	"fmt"

	"strings"

	beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"
)

type KindPorts struct {
	In        map[string]map[string]bool
	Out       map[string]map[string]bool
	Broadcast map[string]map[string]bool
}

func (k KindPorts) Knows(kind string) bool {
	_, ok := k.In[kind]
	return ok
}

func ValidateSpec(spec *TopoSpec, kindPorts KindPorts) error {
	var errs []string

	nodeType := map[string]string{}
	seenID := map[string]bool{}
	for _, n := range spec.Nodes {
		if seenID[n.ID] {

			errs = append(errs, fmt.Sprintf("duplicate node id %q", n.ID))
			continue
		}
		seenID[n.ID] = true
		nodeType[n.ID] = n.Type
		if !kindPorts.Knows(n.Type) {
			errs = append(errs, fmt.Sprintf("node %q: unknown type %q", n.ID, n.Type))
		}
	}

	for _, n := range spec.Nodes {
		if !SafeTreePathComponent(n.ID) {
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
			if base, isMulti := BroadcastBaseName(srcHandle, srcKind, kindPorts.Broadcast); isMulti {
				srcHandle = base
			}
			if !kindPorts.Out[srcKind][srcHandle] {
				errs = append(errs, fmt.Sprintf("edge %q: sourceHandle %q is not an output port on kind %q", e.Label, e.SourceHandle, srcKind))
			}
		}
		tgtKind, tgtKnown := nodeType[e.Target]
		if !tgtKnown {
			errs = append(errs, fmt.Sprintf("edge %q references unknown node id %q as its target", e.Label, e.Target))
		} else if !kindPorts.In[tgtKind][e.TargetHandle] {
			errs = append(errs, fmt.Sprintf("edge %q: targetHandle %q is not an input port on kind %q", e.Label, e.TargetHandle, tgtKind))
		}
	}

	for _, n := range spec.Nodes {
		if n.Data == nil || n.Data.SendRules == nil {
			continue
		}
		for port, raw := range n.Data.SendRules {
			if _, err := beadanimation.ParseSendRule(raw); err != nil {
				errs = append(errs, fmt.Sprintf("node %q port %q: %v", n.ID, port, err))
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("LoadTopology: spec validation failed:\n  %s", strings.Join(errs, "\n  "))
}

func BroadcastBaseName(handle, kind string, kindBroadcastPorts map[string]map[string]bool) (string, bool) {
	if len(handle) == 0 {
		return handle, false
	}
	last := handle[len(handle)-1]
	if last < '0' || last > '9' {
		return handle, false
	}
	base := handle[:len(handle)-1]
	if kindBroadcastPorts[kind][base] {
		return base, true
	}
	return handle, false
}
