package Node

import (
	"context"
	"fmt"
	"math"

	"github.com/dtauraso/wirefold/Categories/Node/rulenode"
)

func EditNode(ctx context.Context, e Edit, rules *rulenode.RuleChannels) {
	if e.Attr == "kindActive" {
		ToggleKindRule(ctx, rules, e.Num)
		return
	}

	kind, ok := nodeAttrEditKinds[e.Attr]
	if !ok {
		return
	}
	edit := rulenode.Edit{Kind: kind}
	if e.Attr == "dragMaxTheta" || e.Attr == "selfDragMaxTheta" {
		if e.X >= 0 {
			radians := e.X * math.Pi
			edit.MaxTheta = &radians
		}
	}
	SendRuleEdit(ctx, rules, e.Num, edit)
}

var nodeAttrEditKinds = map[string]rulenode.EditKind{
	"dragPhi":          rulenode.EditPhiToggle,
	"dragMaxTheta":     rulenode.EditMaxTheta,
	"dragActive":       rulenode.EditActiveToggle,
	"dragR":            rulenode.EditRToggle,
	"selfDragR":        rulenode.EditSelfRToggle,
	"selfDragPhi":      rulenode.EditSelfPhiToggle,
	"selfDragMaxTheta": rulenode.EditSelfMaxTheta,
	"selfDragActive":   rulenode.EditSelfActiveToggle,
}

func SendRuleEdit(ctx context.Context, rules *rulenode.RuleChannels, row int, edit rulenode.Edit) {
	if row < 0 || row >= len(rules.EditsByNodeRow) {
		panic(fmt.Sprintf(
			"sendRuleEdit: node row %d is outside the %d rows the tree declares, so a rule edit names an entity "+
				"the row space has no slot for — the webview and the loaded tree disagree about how many nodes exist",
			row, len(rules.EditsByNodeRow)))
	}
	edits := rules.EditsByNodeRow[row]
	if edits == nil {
		return
	}
	select {
	case edits <- edit:
	case <-ctx.Done():
	}
}

func ToggleKindRule(ctx context.Context, rules *rulenode.RuleChannels, row int) {
	if row < 0 || row >= len(rules.KindTogglesByNodeRow) {
		panic(fmt.Sprintf(
			"kindActive: node row %d is outside the %d rows the tree declares, so a kind-rule toggle names an "+
				"entity the row space has no slot for — the webview and the loaded tree disagree about how many "+
				"nodes exist", row, len(rules.KindTogglesByNodeRow)))
	}
	toggle := rules.KindTogglesByNodeRow[row]
	if toggle == nil {
		return
	}
	select {
	case toggle <- struct{}{}:
	case <-ctx.Done():
	}
}
