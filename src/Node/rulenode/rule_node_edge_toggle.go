package rulenode

import (
	"context"

	"github.com/dtauraso/wirefold/src/jsonpersist"
	"github.com/dtauraso/wirefold/src/Node/nodeactor/nodefiles"
)

type EdgeToggle struct {
	Target string
}

func (r *RuleNode) SeedKindActive(active bool) { r.kindActive = active }

func (r *RuleNode) KindActive() bool { return r.kindActive }

func (r *RuleNode) KindToggleChannel() chan<- struct{} { return r.toggleSelfKind }

func (r *RuleNode) forwardKindToggle(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.toggleSelfKind:
			select {
			case r.kindIn <- struct{}{}:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (r *RuleNode) applyKindToggle() {
	r.kindActive = !r.kindActive
	if r.persistRoot == "" {
		return
	}
	if err := nodefiles.WriteKindRuleActive(r.persistRoot, r.id, r.kindActive); err != nil {
		jsonpersist.LogPersistErr("rulenode", r.id, err)
	}
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
	if err := nodefiles.WriteEdgeRuleActive(r.persistRoot, r.id, t.Target, next); err != nil {
		jsonpersist.LogPersistErr("rulenode", r.id+"To"+t.Target, err)
	}
}
