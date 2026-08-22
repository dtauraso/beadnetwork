package Dispatch

import (
	"context"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"
	NodeKind "github.com/dtauraso/wirefold/src/Node"
	"github.com/dtauraso/wirefold/src/Node/rulenode"
)

func applyRulesHit(ctx context.Context, md *MoveDispatch, h PolarRulesPanel.Hit) {
	switch h.Kind {
	case PolarRulesPanel.HitToggle:
		Panel.ToggleFlag(&md.UI.PN, "nodeRules")
		md.Persist.Panels().Schedule(md.UI.PN)
	case PolarRulesPanel.HitShared:
		md.UI.ToggleSharedRow(h.NodeRow)
	case PolarRulesPanel.HitMenuRow:
		if h.NodeRow < 0 {
			for _, n := range md.UI.RuleNodes {
				sendRuleEdit(ctx, md, int(n.Row), rulenode.Edit{Kind: rulenode.EditActiveToggle})
			}
			break
		}
		sendRuleEdit(ctx, md, int(h.NodeRow), rulenode.Edit{Kind: rulenode.EditActiveToggle})
	case PolarRulesPanel.HitCheck:
		NodeKind.ApplyRuleCheck(ctx, h, &md.Rules)
	case PolarRulesPanel.HitValue:
		if !NodeKind.ApplyRuleValue(ctx, h, &md.Rules) {
			md.UI.StartThetaDraft(h.NodeRow, h.Value == PolarRulesPanel.ValSelfTheta)
		}
	}
	md.UI.EmitViewFrame(nil)
}

func applyRuleKey(ctx context.Context, md *MoveDispatch, key string) {
	if c := md.UI.RuleKey(key); c.Commit {
		NodeKind.CommitMaxTheta(ctx, &md.Rules, c.NodeRow, c.Self, c.Turns)
	}
}
