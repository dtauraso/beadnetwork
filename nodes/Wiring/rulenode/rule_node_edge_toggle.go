package rulenode

import (
	"context"

	"github.com/dtauraso/wirefold/nodes/Wiring/edgefile"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

type EdgeToggle struct {
	Target string
}

func (r *RuleNode) SeedEdgeActive(target string, active bool) {
	r.edgeActive[target] = active
	if _, made := r.toggleSelfToPeer[target]; !made {
		r.toggleSelfToPeer[target] = make(chan struct{}, 4)
	}
}

func (r *RuleNode) EdgeToggleChannel(target string) chan<- struct{} {
	if selfToTarget, made := r.toggleSelfToPeer[target]; made {
		return selfToTarget
	}
	selfToTarget := make(chan struct{}, 4)
	r.toggleSelfToPeer[target] = selfToTarget
	return selfToTarget
}

func (r *RuleNode) EdgeActive(target string) bool {
	active, seeded := r.edgeActive[target]
	return !seeded || active
}

func (r *RuleNode) forwardToggle(ctx context.Context, target string, toggle chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-toggle:
			select {
			case r.toggleIn <- EdgeToggle{Target: target}:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (r *RuleNode) applyEdgeToggle(t EdgeToggle) {
	next := !r.EdgeActive(t.Target)
	r.edgeActive[t.Target] = next
	if r.persistRoot == "" {
		return
	}
	label := r.id + "To" + t.Target
	if err := edgefile.WriteEdgeRuleActive(r.persistRoot, r.id, label, next); err != nil {
		jsonpersist.LogPersistErr("rulenode", label, err)
	}
}
