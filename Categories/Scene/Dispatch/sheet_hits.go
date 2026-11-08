package Dispatch

import (
	"context"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"
)

func applySheetHit(ctx context.Context, md *MoveDispatch, h Sheet.Hit) {
	sh := &md.UI.Sheet
	switch h.Part {
	case Sheet.PartPill:
		sh.Toggle(h.Label)
	case Sheet.PartGrab:
		sh.StartResize(h.Label, h.DY)
	case Sheet.PartBody:
		if h.Cell.Act.Op == Sheet.OpDockFold {
			sh.ToggleFold(h.Cell.Act.Node)
		} else if h.Cell.Act.Op == Sheet.OpDockTab {
			sh.SelectTab(h.Cell.Act.I)
			sh.StartTabDrag(h.Cell.Act.Node)
		} else if h.Cell.Typed {
			Sheet.StartDraft(&sh.Edit, h.Cell.Act)
		} else {
			runSheetAct(ctx, md, h.Cell.Act, 0, nil)
		}
	case Sheet.PartUp, Sheet.PartDown:
		if !h.Cell.Dim {
			runSheetAct(ctx, md, h.Cell.ArrowAct(), map[Sheet.Part]int{Sheet.PartUp: 1, Sheet.PartDown: -1}[h.Part], nil)
		}
	}
	md.UI.EmitViewFrame(nil)
}

func applySheetKey(ctx context.Context, md *MoveDispatch, key string) bool {
	act, v, ok, err := Sheet.Key(&md.UI.Sheet.Edit, key)
	if err != nil {
		md.UI.Builder.Refused = err.Error()
	}
	if ok {
		runSheetAct(ctx, md, act, 0, &v)
	}
	md.UI.EmitViewFrame(nil)
	return err != nil || (ok && md.UI.Builder.Refused != "")
}

func runSheetAct(ctx context.Context, md *MoveDispatch, a Sheet.Act, sign int, typed *float64) {
	md.UI.Builder.Refused = ""
	if Builder.IsStepEdit(a) {
		md.UI.Builder.Refused = md.UI.Builder.StepEdit(a, sign, typed)
		md.persistStep()
		if md.UI.Builder.Refused == "" {
			md.syncStepPicks(ctx, a.I)
			md.sendStepLink(ctx, Builder.StepEdge{Step: a.I, From: a.Node, To: a.To})
		}
		return
	}
	if Builder.IsStepHead(a) {
		md.runStepHead(ctx, a, sign)
		return
	}
	if runSheetStructure(ctx, md, a, sign, typed) || runSheetSequence(md, a) {
		return
	}
	switch a.Op {
	case Builder.OpStepPick:
		if md.UI.Builder.Refused = md.UI.Builder.ToggleStepPick(a.I, a.Node, a.To); md.UI.Builder.Refused == "" {
			md.persistStep()
			md.syncStepPicks(ctx, a.I)
		}
	case Builder.OpStepFold:
		md.UI.Builder.ToggleStepFold(a.I, a.Node)
	case Builder.OpSeqFold:
		md.UI.Builder.SeqFolded = !md.UI.Builder.SeqFolded
	case Builder.OpItem, Builder.OpMethod, Builder.OpCard:
		setCards(md, a)
	}
}

func (md *MoveDispatch) Settle(ctx context.Context) { md.stepFrom(ctx, "", Links.KBoth, 0) }

func (md *MoveDispatch) stepFrom(ctx context.Context, from, k string, sign int) {
	m := &md.UI.Builder
	m.Step.Round++
	h := clock.HumanTick(m.TickMs)
	for _, id := range m.Nodes {
		md.sendLink(ctx, id, Links.StepAll{Round: m.Step.Round, H: h, K: k, Sign: sign, From: from, LockAll: m.StepOf(from).Locks(), FixedR: m.ManyIn()})
	}
}

func setCards(md *MoveDispatch, a Sheet.Act) {
	m := &md.UI.Builder
	c := &m.Cards
	switch a.Op {
	case Builder.OpItem:
		c.Item = a.Node
	case Builder.OpMethod:
		if c.Methods == nil {
			c.Methods = map[string]bool{}
		}
		c.Methods[a.K] = !c.Methods[a.K]
	case Builder.OpCard:
		c.Card = m.CardShown() + a.Sign
	}
	c.Card = m.CardShown()
	if m.CardsDir != "" {
		if err := Builder.WriteCards(m.CardsDir, *c); err != nil {
			View.LogPersistErr("sheet_cards_persist", m.CardsDir, err)
		}
	}
}
