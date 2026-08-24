package Node

import (
	"context"
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/PolarRulesPanel"
	edge "github.com/dtauraso/beadnetwork/Categories/Node/Edge"
)

func ApplyRuleCheck(ctx context.Context, h PolarRulesPanel.Hit, rules *RuleChannels) {
	switch h.Check {
	case PolarRulesPanel.CheckNodeDrag:
		SendRuleEdit(ctx, rules, int(h.NodeRow), RuleEdit{Kind: EditActiveToggle})
	case PolarRulesPanel.CheckSelfDrag:
		SendRuleEdit(ctx, rules, int(h.NodeRow), RuleEdit{Kind: EditSelfActiveToggle})
	case PolarRulesPanel.CheckKindRule:
		ToggleKindRule(ctx, rules, int(h.NodeRow))
	case PolarRulesPanel.CheckEdgeDrag:
		edge.ToggleDragActive(ctx, int(h.EdgeRow), rules.TogglesByEdgeRow)
	}
}

var ruleValueEdits = map[PolarRulesPanel.ValueKind]EditKind{
	PolarRulesPanel.ValSelfR:   EditSelfRToggle,
	PolarRulesPanel.ValSelfPhi: EditSelfPhiToggle,
	PolarRulesPanel.ValDragR:   EditRToggle,
	PolarRulesPanel.ValDragPhi: EditPhiToggle,
}

func ApplyRuleValue(ctx context.Context, h PolarRulesPanel.Hit, rules *RuleChannels) bool {
	kind, ok := ruleValueEdits[h.Value]
	if !ok {
		return false
	}
	SendRuleEdit(ctx, rules, int(h.NodeRow), RuleEdit{Kind: kind})
	return true
}

func CommitMaxTheta(ctx context.Context, rules *RuleChannels, row int32, self bool, turns float64) {
	var maxTheta *float64
	if turns >= 0 {
		radians := turns * math.Pi
		maxTheta = &radians
	}
	kind := EditMaxTheta
	if self {
		kind = EditSelfMaxTheta
	}
	SendRuleEdit(ctx, rules, int(row), RuleEdit{Kind: kind, MaxTheta: maxTheta})
}
