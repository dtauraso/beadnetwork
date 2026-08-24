package Ports

import "github.com/dtauraso/beadnetwork/Categories/Scene/Topology"

func KindPortSets() Topology.KindPorts {
	sets := Topology.KindPorts{
		In:        map[string]map[string]bool{},
		Out:       map[string]map[string]bool{},
		Broadcast: map[string]map[string]bool{},
	}
	for kind, ports := range KindPorts {
		ins := map[string]bool{}
		outs := map[string]bool{}
		broadcasts := map[string]bool{}
		for _, p := range ports {
			switch p.Dir {
			case PortIn:
				ins[p.Name] = true
			case PortOut:
				outs[p.Name] = true
			case PortBroadcast:
				broadcasts[p.Name] = true
				outs[p.Name] = true
			}
		}
		sets.In[kind] = ins
		sets.Out[kind] = outs
		sets.Broadcast[kind] = broadcasts
	}
	return sets
}
