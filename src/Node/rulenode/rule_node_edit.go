package rulenode

import (
	"fmt"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"

	"github.com/dtauraso/wirefold/src/valuefile"
	"github.com/dtauraso/wirefold/src/Node/nodeactor/nodefiles"
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
	if r.persistRoot == "" {
		return
	}
	if err := nodefiles.WriteDragRule(r.persistRoot, r.id, r.rule); err != nil {
		valuefile.LogPersistErr("rulenode", r.id, err)
	}
}

func (r *RuleNode) persistSelfRule() {
	if r.persistRoot == "" {
		return
	}
	if err := nodefiles.WriteSelfDragRule(r.persistRoot, r.id, r.selfRule); err != nil {
		valuefile.LogPersistErr("rulenode", r.id, err)
	}
}

func (r *RuleNode) persistSelfActive() {
	if r.persistRoot == "" {
		return
	}
	if err := nodefiles.WriteSelfRuleActive(r.persistRoot, r.id, r.selfActive); err != nil {
		valuefile.LogPersistErr("rulenode", r.id, err)
	}
}

func (r *RuleNode) persistActive() {
	if r.persistRoot == "" {
		return
	}
	if err := nodefiles.WriteDragActive(r.persistRoot, r.id, r.active); err != nil {
		valuefile.LogPersistErr("rulenode", r.id, err)
	}
}
