package rulenode

import (
	"fmt"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/PolarRulesPanel"
)

func (r *RuleNode) applyEdit(e Edit) {
	switch e.Kind {
	case EditActiveToggle:
		r.active = !r.active
		r.persistActive()
	case EditPhiToggle:
		var next PolarRulesPanel.DragRule
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
		var next PolarRulesPanel.DragRule
		if r.rule != nil {
			next = *r.rule
		}
		next.MaxTheta = e.MaxTheta
		r.rule = &next
		r.persistRule()
	case EditRToggle:
		var next PolarRulesPanel.DragRule
		if r.rule != nil {
			next = *r.rule
		}
		if next.R != nil {
			next.R = nil
		} else {
			zero := 0.0
			next.R = &zero
		}
		r.rule = &next
		r.persistRule()
	case EditSelfRToggle:
		var next PolarRulesPanel.DragRule
		if r.selfRule != nil {
			next = *r.selfRule
		}
		if next.R != nil {
			next.R = nil
		} else {
			zero := 0.0
			next.R = &zero
		}
		r.selfRule = &next
		r.persistSelfRule()
	case EditSelfActiveToggle:
		r.selfActive = !r.selfActive
		r.persistSelfActive()
	case EditSelfPhiToggle:
		var next PolarRulesPanel.DragRule
		if r.selfRule != nil {
			next = *r.selfRule
		}
		if next.Phi != nil {
			next.Phi = nil
		} else {
			zero := 0.0
			next.Phi = &zero
		}
		r.selfRule = &next
		r.persistSelfRule()
	case EditSelfMaxTheta:
		var next PolarRulesPanel.DragRule
		if r.selfRule != nil {
			next = *r.selfRule
		}
		next.MaxTheta = e.MaxTheta
		r.selfRule = &next
		r.persistSelfRule()
	default:
		panic(fmt.Sprintf(
			"rulenode.applyEdit: node %q was sent edit kind %d, which no case handles, so a rule edit would be "+
				"silently dropped after crossing the bridge", r.id, e.Kind))
	}
	r.mesh.SetSelfRuleKey(PolarRulesPanel.KeyOf(r.rule))
	r.mesh.BroadcastRule(r.id)
}

func (r *RuleNode) persistRule() {
	if r.persist.Rule == nil {
		return
	}
	if err := r.persist.Rule(r.rule); err != nil {
		LogPersistErr("rulenode", r.id, err)
	}
}

func (r *RuleNode) persistSelfRule() {
	if r.persist.SelfRule == nil {
		return
	}
	if err := r.persist.SelfRule(r.selfRule); err != nil {
		LogPersistErr("rulenode", r.id, err)
	}
}

func (r *RuleNode) persistSelfActive() {
	if r.persist.SelfActive == nil {
		return
	}
	if err := r.persist.SelfActive(r.selfActive); err != nil {
		LogPersistErr("rulenode", r.id, err)
	}
}

func (r *RuleNode) persistActive() {
	if r.persist.DragActive == nil {
		return
	}
	if err := r.persist.DragActive(r.active); err != nil {
		LogPersistErr("rulenode", r.id, err)
	}
}
