package rulenode

import (
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/nodeactor/nodefiles"
	"github.com/dtauraso/wirefold/nodes/Wiring/rulemsg"
)

func (r *RuleNode) applyEdit(e Edit) {
	switch e.Kind {
	case EditActiveToggle:
		r.active = !r.active
		r.persistActive()
	case EditPhiToggle:
		var next polar.DragRule
		if r.rule != nil {
			next = *r.rule
		}
		if next.Phi != nil {
			next.Phi = nil
		} else {
			zero := 0.0
			next.Phi = &zero
		}
		r.rule = &next
		r.persistRule()
	case EditMaxTheta:
		var next polar.DragRule
		if r.rule != nil {
			next = *r.rule
		}
		next.MaxTheta = e.MaxTheta
		r.rule = &next
		r.persistRule()
	default:
		panic(fmt.Sprintf(
			"rulenode.applyEdit: node %q was sent edit kind %d, which no case handles, so a rule edit would be "+
				"silently dropped after crossing the bridge", r.id, e.Kind))
	}
	r.mesh.SetSelfRuleKey(rulemsg.KeyOf(r.rule))
	r.mesh.BroadcastRule(r.id)
}

func (r *RuleNode) persistRule() {
	if r.persistRoot == "" {
		return
	}
	if err := nodefiles.WriteDragRule(r.persistRoot, r.id, r.rule); err != nil {
		jsonpersist.LogPersistErr("rulenode", r.id, err)
	}
}

func (r *RuleNode) persistActive() {
	if r.persistRoot == "" {
		return
	}
	if err := nodefiles.WriteDragActive(r.persistRoot, r.id, r.active); err != nil {
		jsonpersist.LogPersistErr("rulenode", r.id, err)
	}
}
