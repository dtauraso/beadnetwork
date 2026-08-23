package Node

import (
	"context"
	"fmt"
	"math"
)

func EditNode(ctx context.Context, e Edit, rules *RuleChannels) {
	if e.Attr == "kindActive" {
		ToggleKindRule(ctx, rules, e.Num)
		return
	}

	kind, ok := nodeAttrEditKinds[e.Attr]
	if !ok {
		return
	}
	edit := RuleEdit{Kind: kind}
	if e.Attr == "dragMaxTheta" || e.Attr == "selfDragMaxTheta" {
		if e.X >= 0 {
			radians := e.X * math.Pi
			edit.MaxTheta = &radians
		}
	}
	SendRuleEdit(ctx, rules, e.Num, edit)
}

var nodeAttrEditKinds = map[string]EditKind{
	"dragPhi":          EditPhiToggle,
	"dragMaxTheta":     EditMaxTheta,
	"dragActive":       EditActiveToggle,
	"dragR":            EditRToggle,
	"selfDragR":        EditSelfRToggle,
	"selfDragPhi":      EditSelfPhiToggle,
	"selfDragMaxTheta": EditSelfMaxTheta,
	"selfDragActive":   EditSelfActiveToggle,
}

func SendRuleEdit(ctx context.Context, rules *RuleChannels, row int, edit RuleEdit) {
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

func ToggleKindRule(ctx context.Context, rules *RuleChannels, row int) {
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

func ApplyUpdate(ctx context.Context, attr byte, payload []byte, rules *RuleChannels) {
	e, ok := DecodeUpdate(payload, attr)
	if !ok {
		return
	}
	EditNode(ctx, e, rules)
}
