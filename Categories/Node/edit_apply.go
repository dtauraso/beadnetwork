package Node

import (
	"context"
	"fmt"
	"math"

	"github.com/dtauraso/wirefold/Categories/Input/Stdin"
	"github.com/dtauraso/wirefold/Categories/Node/rulechans"
	"github.com/dtauraso/wirefold/Categories/Node/rulenode"
)

func EditNode(ctx context.Context, msg Stdin.StdinMsg, rules *rulechans.RuleChannels) {
	if msg.Attr == "kindActive" {
		ToggleKindRule(ctx, rules, msg.Num)
		return
	}

	kind, ok := nodeAttrEditKinds[msg.Attr]
	if !ok {
		return
	}
	edit := rulenode.Edit{Kind: kind}
	if msg.Attr == "dragMaxTheta" || msg.Attr == "selfDragMaxTheta" {
		if msg.X >= 0 {
			radians := msg.X * math.Pi
			edit.MaxTheta = &radians
		}
	}
	SendRuleEdit(ctx, rules, msg.Num, edit)
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

func SendRuleEdit(ctx context.Context, rules *rulechans.RuleChannels, row int, edit rulenode.Edit) {
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

func ToggleKindRule(ctx context.Context, rules *rulechans.RuleChannels, row int) {
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
