package rulenode

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgefile"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

type EdgeToggle struct {
	Target string
	Active bool
}

func (r *RuleNode) SeedEdgeActive(target string, active bool) {
	r.edgeActive[target] = active
	if _, made := r.toggleSelfToPeer[target]; !made {
		r.toggleSelfToPeer[target] = make(chan bool, 4)
	}
}

func (r *RuleNode) EdgeToggleChannel(target string) chan<- bool {
	if selfToTarget, made := r.toggleSelfToPeer[target]; made {
		return selfToTarget
	}
	selfToTarget := make(chan bool, 4)
	r.toggleSelfToPeer[target] = selfToTarget
	return selfToTarget
}

func (r *RuleNode) EdgeActive(target string) bool {
	active, seeded := r.edgeActive[target]
	return !seeded || active
}

func (r *RuleNode) forwardToggle(ctx context.Context, target string, toggle chan bool) {
	for {
		select {
		case <-ctx.Done():
			return
		case active := <-toggle:
			select {
			case r.toggleIn <- EdgeToggle{Target: target, Active: active}:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (r *RuleNode) applyEdgeToggle(t EdgeToggle) {
	r.edgeActive[t.Target] = t.Active
	if r.persistRoot == "" {
		return
	}
	label := r.id + "To" + t.Target
	if err := edgefile.WriteEdgeRuleActive(r.persistRoot, r.id, label, t.Active); err != nil {
		jsonpersist.LogPersistErr("rulenode", label, err)
	}
}
