package scenebuild

import (
	beadanimation "github.com/dtauraso/wirefold/Categories/Node/BeadAnimation"
)

type EdgeWiring struct {
	Inbound        map[string]map[string]string
	Outbound       map[string]map[string][]string
	OutboundHandle map[string]map[string][]string

	DestRun map[string]*beadanimation.BeadLine
	EdgeRun map[string]*beadanimation.BeadLine
}

func (w EdgeWiring) BindPorts(pb *PortBindings, n Node, ports []PortSpec) {
	for _, port := range ports {
		switch port.Dir {
		case PortIn:
			if dk, ok := w.Inbound[n.ID][port.Name]; ok {
				pb.SetSinglePaced(port.Name, w.DestRun[dk])
			}

		case PortOut:
			labels := w.Outbound[n.ID][port.Name]
			if len(labels) > 0 {
				lbl := labels[0]
				pb.SetSinglePacedRule(port.Name, w.EdgeRun[lbl], NodeSendRule(n, port.Name), lbl)
			}

		case PortBroadcast:
			labels := w.Outbound[n.ID][port.Name]
			handles := w.OutboundHandle[n.ID][port.Name]
			for i, lbl := range labels {
				handle := port.Name
				if i < len(handles) {
					handle = handles[i]
				}
				pb.AppendBroadcastWithHandle(port.Name, handle, w.EdgeRun[lbl], NodeSendRule(n, handle), lbl)
			}
		}
	}
}
