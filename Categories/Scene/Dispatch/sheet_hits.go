package Dispatch

import (
	"context"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"
	Flags "github.com/dtauraso/beadnetwork/Categories/Scene/View/Flags"
)

func applySheetHit(ctx context.Context, md *MoveDispatch, h Sheet.Hit) {
	sh := &md.UI.Sheet
	if op := h.Cell.Act.Op; h.Part != Sheet.PartBody || op != Builder.OpSeqFold && op != Builder.OpSeqOpen {
		md.UI.Builder.SeqListOpen = false
	}
	if h.Part == Sheet.PartBody || h.Part == Sheet.PartUp || h.Part == Sheet.PartDown {
		md.UI.Builder.Touch(h.Cell.ArrowAct())
	}
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
		} else if h.Cell.Act.Op == Builder.OpStepShow {
			md.UI.Builder.DragStep = h.Cell.Act.I
		} else if h.Cell.Typed {
			Sheet.StartDraft(&sh.Edit, h.Cell)
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

func endStepDrag(md *MoveDispatch, x, y float64) bool {
	m := &md.UI.Builder
	from := m.DragStep
	if from == 0 {
		return false
	}
	m.DragStep, m.Refused = 0, ""
	h := md.UI.PanelLayout().Sheet.Hit(x, y)
	if to := h.Cell.Act.I; h.Part == Sheet.PartBody && h.Cell.Act.Op == Builder.OpStepShow && to != from {
		refused, err := m.MoveStep(from, to)
		m.Refused = refused
		if err != nil {
			View.LogPersistErr("sheet_step_move", m.Root, err)
		}
	} else {
		m.ShowStep(Sheet.Act{Op: Builder.OpStepShow, I: from}, 0)
	}
	md.UI.EmitViewFrame(nil)
	return true
}

func applySheetKey(ctx context.Context, md *MoveDispatch, key string) bool {
	act, t, ok, err := Sheet.Key(&md.UI.Sheet.Edit, key)
	if err != nil {
		md.UI.Builder.Refused = err.Error()
	}
	if ok && act.Op == Builder.OpStepName {
		md.UI.Builder.Refused = md.UI.Builder.NameStep(act.I, t.Text)
	} else if ok {
		runSheetAct(ctx, md, act, 0, &t.V)
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
			md.sFitsStarts(ctx)
		}
		return
	}
	if Builder.IsStepHead(a) {
		md.runStepHead(ctx, a, sign)
		return
	}
	if Builder.IsStepDirOp(a) {
		refused, err := md.UI.Builder.StepDirOp(a)
		md.UI.Builder.Refused = refused
		if err != nil {
			View.LogPersistErr("sheet_step_dir", md.UI.Builder.Root, err)
		}
		return
	}
	if a.Op == Builder.OpSeqOpen {
		md.UI.Builder.SeqListOpen = false
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
		md.UI.Builder.SeqListOpen = !md.UI.Builder.SeqListOpen
	case Builder.OpStepShow:
		md.UI.Builder.ShowStep(a, sign)
	case Builder.OpFolderFold:
		md.UI.Builder.ToggleFolder(a.Node)
	case Builder.OpNetFold:
		md.UI.Builder.NetworkOpen = !md.UI.Builder.NetworkOpen
	case Builder.OpItem, Builder.OpMethod, Builder.OpCard:
		setCards(md, a)
	case Pills.OpOverlay:
		Flags.ToggleFlag(&md.UI.OV, &md.ChannelVectorsOn, &md.UI, a.K)
		md.UI.PersistOverlays(md.UI.OV)
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
