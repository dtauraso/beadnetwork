package Node

import "github.com/dtauraso/wirefold/Categories/Chrome/Panels/PolarRulesPanel"

type Persist struct {
	Rule func(rule *PolarRulesPanel.DragRule) error

	SelfRule func(rule *PolarRulesPanel.DragRule) error

	SelfActive func(active bool) error

	DragActive func(active bool) error

	KindActive func(active bool) error

	EdgeActive func(target string, active bool) error
}

func (r *RuleNode) SetPersist(p Persist) { r.persist = p }
