package Dispatch

import (
	"context"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/PolarRulesPanel"
	NodeKind "github.com/dtauraso/beadnetwork/Categories/Node"
)

func applyRulesHit(ctx context.Context, md *MoveDispatch, h PolarRulesPanel.Hit) {
	switch h.Kind {
	case PolarRulesPanel.HitToggle:
		Panel.ToggleFlag(&md.UI.PN, "nodeRules")
		md.UI.PersistPanels(md.UI.PN)
	case PolarRulesPanel.HitShared:
		PolarRulesPanel.ToggleSharedRow(&md.UI.Rules.SharedRow, h.NodeRow)
	case PolarRulesPanel.HitMenuRow:
		if h.NodeRow < 0 {
			for _, n := range md.UI.Rules.Nodes {
				sendRuleEdit(ctx, md, int(n.Row), NodeKind.RuleEdit{Kind: NodeKind.EditActiveToggle})
			}
			break
		}
		sendRuleEdit(ctx, md, int(h.NodeRow), NodeKind.RuleEdit{Kind: NodeKind.EditActiveToggle})
	case PolarRulesPanel.HitCheck:
		NodeKind.ApplyRuleCheck(ctx, h, &md.Rules)
	case PolarRulesPanel.HitValue:
		if !NodeKind.ApplyRuleValue(ctx, h, &md.Rules) {
			PolarRulesPanel.StartThetaDraft(&md.UI.Rules.Edit, h.NodeRow, h.Value == PolarRulesPanel.ValSelfTheta)
		}
	}
	md.UI.EmitViewFrame(nil)
}

func applyRuleKey(ctx context.Context, md *MoveDispatch, key string) {
	if c := PolarRulesPanel.RuleKey(&md.UI.Rules.Edit, key, md.redraw); c.Commit {
		NodeKind.CommitMaxTheta(ctx, &md.Rules, c.NodeRow, c.Self, c.Turns)
	}
}
