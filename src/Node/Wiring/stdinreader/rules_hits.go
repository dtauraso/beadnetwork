package stdinreader

import (
	"context"
	"math"

	"github.com/dtauraso/wirefold/src/Node/Wiring/dispatch"
	"github.com/dtauraso/wirefold/src/Node/Wiring/rulenode"
	"github.com/dtauraso/wirefold/src/Node/Wiring/rulespanel"
	"github.com/dtauraso/wirefold/src/Chrome/OverlaysDropdown"
)

func applyRulesHit(ctx context.Context, md *dispatch.MoveDispatch, h rulespanel.Hit) {
	switch h.Kind {
	case rulespanel.HitToggle:
		if fn, ok := OverlaysDropdown.PanelToggles["nodeRules"]; ok {
			fn(&md.UI.PN)
			md.Persist.Panels().Schedule(md.UI.PN)
		}
	case rulespanel.HitShared:
		if md.UI.RuleSharedRow == h.NodeRow {
			md.UI.RuleSharedRow = -1
		} else {
			md.UI.RuleSharedRow = h.NodeRow
		}
	case rulespanel.HitMenuRow:
		if h.NodeRow < 0 {
			for _, n := range md.UI.RuleNodes {
				sendRuleEdit(ctx, md, int(n.Row), rulenode.Edit{Kind: rulenode.EditActiveToggle})
			}
			break
		}
		sendRuleEdit(ctx, md, int(h.NodeRow), rulenode.Edit{Kind: rulenode.EditActiveToggle})
	case rulespanel.HitCheck:
		applyRuleCheck(ctx, md, h)
	case rulespanel.HitValue:
		applyRuleValue(ctx, md, h)
	}
	md.UI.EmitViewFrame(nil)
}

func applyRuleCheck(ctx context.Context, md *dispatch.MoveDispatch, h rulespanel.Hit) {
	switch h.Check {
	case rulespanel.CheckNodeDrag:
		sendRuleEdit(ctx, md, int(h.NodeRow), rulenode.Edit{Kind: rulenode.EditActiveToggle})
	case rulespanel.CheckSelfDrag:
		sendRuleEdit(ctx, md, int(h.NodeRow), rulenode.Edit{Kind: rulenode.EditSelfActiveToggle})
	case rulespanel.CheckKindRule:
		toggleKindRule(ctx, md, int(h.NodeRow))
	case rulespanel.CheckEdgeDrag:
		toggleEdgeDragActive(ctx, md, h.EdgeRow)
	}
}

func applyRuleValue(ctx context.Context, md *dispatch.MoveDispatch, h rulespanel.Hit) {
	switch h.Value {
	case rulespanel.ValSelfR:
		sendRuleEdit(ctx, md, int(h.NodeRow), rulenode.Edit{Kind: rulenode.EditSelfRToggle})
	case rulespanel.ValSelfPhi:
		sendRuleEdit(ctx, md, int(h.NodeRow), rulenode.Edit{Kind: rulenode.EditSelfPhiToggle})
	case rulespanel.ValDragR:
		sendRuleEdit(ctx, md, int(h.NodeRow), rulenode.Edit{Kind: rulenode.EditRToggle})
	case rulespanel.ValDragPhi:
		sendRuleEdit(ctx, md, int(h.NodeRow), rulenode.Edit{Kind: rulenode.EditPhiToggle})
	case rulespanel.ValSelfTheta, rulespanel.ValDragTheta:
		md.UI.RuleEdit = rulespanel.Edit{
			Active:  true,
			NodeRow: h.NodeRow,
			Self:    h.Value == rulespanel.ValSelfTheta,
			Draft:   "1/2",
		}
	}
}

func applyRuleKey(ctx context.Context, md *dispatch.MoveDispatch, key string) {
	e := &md.UI.RuleEdit
	if !e.Active {
		return
	}
	switch key {
	case "Escape":
		*e = rulespanel.Edit{}
	case "Enter":
		if turns, ok := rulespanel.ParsePiDraft(e.Draft); ok {
			commitMaxTheta(ctx, md, e.NodeRow, e.Self, turns)
		}
		*e = rulespanel.Edit{}
	case "Backspace":
		if len(e.Draft) > 0 {
			e.Draft = e.Draft[:len(e.Draft)-1]
		}
	default:
		if len(key) == 1 {
			e.Draft += key
		}
	}
	md.UI.EmitViewFrame(nil)
}

func toggleKindRule(ctx context.Context, md *dispatch.MoveDispatch, row int) {
	if row < 0 || row >= len(md.Rules.KindTogglesByNodeRow) {
		return
	}
	sendToggle(ctx, md.Rules.KindTogglesByNodeRow[row])
}

func toggleEdgeDragActive(ctx context.Context, md *dispatch.MoveDispatch, row int32) {
	if row < 0 || int(row) >= len(md.Rules.TogglesByEdgeRow) {
		return
	}
	sendToggle(ctx, md.Rules.TogglesByEdgeRow[row])
}

func sendToggle(ctx context.Context, ch chan<- struct{}) {
	if ch == nil {
		return
	}
	select {
	case ch <- struct{}{}:
	case <-ctx.Done():
	}
}

func commitMaxTheta(ctx context.Context, md *dispatch.MoveDispatch, row int32, self bool, turns float64) {
	var maxTheta *float64
	if turns >= 0 {
		radians := turns * math.Pi
		maxTheta = &radians
	}
	kind := rulenode.EditMaxTheta
	if self {
		kind = rulenode.EditSelfMaxTheta
	}
	sendRuleEdit(ctx, md, int(row), rulenode.Edit{Kind: kind, MaxTheta: maxTheta})
}
