package Node

import (
	"context"
	"math"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"
	edge "github.com/dtauraso/wirefold/src/Node/Edge"
	"github.com/dtauraso/wirefold/src/Node/rulechans"
	"github.com/dtauraso/wirefold/src/Node/rulenode"
)

func ApplyRuleCheck(ctx context.Context, h PolarRulesPanel.Hit, rules *rulechans.RuleChannels) {
	switch h.Check {
	case PolarRulesPanel.CheckNodeDrag:
		SendRuleEdit(ctx, rules, int(h.NodeRow), rulenode.Edit{Kind: rulenode.EditActiveToggle})
	case PolarRulesPanel.CheckSelfDrag:
		SendRuleEdit(ctx, rules, int(h.NodeRow), rulenode.Edit{Kind: rulenode.EditSelfActiveToggle})
	case PolarRulesPanel.CheckKindRule:
		ToggleKindRule(ctx, rules, int(h.NodeRow))
	case PolarRulesPanel.CheckEdgeDrag:
		edge.ToggleDragActive(ctx, int(h.EdgeRow), rules.TogglesByEdgeRow)
	}
}

var ruleValueEdits = map[PolarRulesPanel.ValueKind]rulenode.EditKind{
	PolarRulesPanel.ValSelfR:   rulenode.EditSelfRToggle,
	PolarRulesPanel.ValSelfPhi: rulenode.EditSelfPhiToggle,
	PolarRulesPanel.ValDragR:   rulenode.EditRToggle,
	PolarRulesPanel.ValDragPhi: rulenode.EditPhiToggle,
}

func ApplyRuleValue(ctx context.Context, h PolarRulesPanel.Hit, rules *rulechans.RuleChannels) bool {
	kind, ok := ruleValueEdits[h.Value]
	if !ok {
		return false
	}
	SendRuleEdit(ctx, rules, int(h.NodeRow), rulenode.Edit{Kind: kind})
	return true
}

func CommitMaxTheta(ctx context.Context, rules *rulechans.RuleChannels, row int32, self bool, turns float64) {
	var maxTheta *float64
	if turns >= 0 {
		radians := turns * math.Pi
		maxTheta = &radians
	}
	kind := rulenode.EditMaxTheta
	if self {
		kind = rulenode.EditSelfMaxTheta
	}
	SendRuleEdit(ctx, rules, int(row), rulenode.Edit{Kind: kind, MaxTheta: maxTheta})
}
