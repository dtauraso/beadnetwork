package Dispatch

import (
	"context"
	"fmt"

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
	case Sheet.PartBody:
		if h.Cell.Typed {
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
	m := md.UI.Builder
	md.UI.Builder.Refused = ""
	if Builder.IsStepEdit(a) {
		md.UI.Builder.Refused = md.UI.Builder.StepEdit(a, sign, typed)
		md.persistStep()
		if l := md.UI.Builder.StepLinks[Builder.StepEdge{Step: a.I, From: a.Node, To: a.To}]; l != nil && md.UI.Builder.Refused == "" {
			md.syncStepPicks(ctx, a.I)
			md.sendLink(ctx, a.To, Links.SetLink{From: a.Node, Start: l.Start, Delta: l.Delta, Increment: l.Increment})
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
	case Builder.OpLinkStart:
		l, ok := m.Link(a.Node, a.To)
		if !ok {
			return
		}
		v := float64(sign) + map[string]float64{Links.KPhi: l.Start.Phi, Links.KTheta: l.Start.Theta, Links.KR: l.Start.R}[a.K]
		if s := m.ScaleOf(a.To); a.K != Links.KR {
			was := map[string]float64{Links.KPhi: l.Start.Phi, Links.KTheta: l.Start.Theta}[a.K]
			v = Links.Wrap(s, was+float64(sign))
		}
		if typed != nil {
			v = *typed
		}
		if a.K == Links.KR && md.refuseFixedR(a, fmt.Sprintf("r stays at %v", l.Start.R)) {
			return
		}
		if a.K == Links.KR && (v < 0 || v+float64(l.Delta.R) < 0) {
			md.UI.Builder.Refused = fmt.Sprintf("%s→%s r %v: neither r nor r + Δr can be negative, so r stays at %v", a.Node, a.To, v, l.Start.R)
			return
		}
		md.sendLink(ctx, a.To, Links.SetStart{From: a.Node, K: a.K, V: v})
	case Builder.OpLinkPress:
		if sign == 0 {
			sign = a.Sign
		}
		if l, ok := m.Link(a.Node, a.To); ok && a.K == Links.KR && md.refuseFixedR(a, fmt.Sprintf("Δr stays at %d", l.Delta.R)) {
			return
		}
		if sign != 0 {
			md.sendLink(ctx, a.To, Links.Press{From: a.Node, K: a.K, Sign: sign})
		}
	case Builder.OpStepPick:
		if md.UI.Builder.Refused = md.UI.Builder.ToggleStepPick(a.I, a.Node, a.To); md.UI.Builder.Refused == "" {
			md.persistStep()
			md.syncStepPicks(ctx, a.I)
		}
	case Builder.OpStepFold:
		md.UI.Builder.ToggleStepFold(a.I, a.Node)
	case Builder.OpIncrement:
		md.setIncrement(ctx, a, sign, typed)
	case Builder.OpItem, Builder.OpMethod, Builder.OpCard:
		setCards(md, a)
	case Builder.OpAllLock:
		st := md.UI.Builder.NodeStep(a.Node)
		way, on := st.At(a.K)
		switch {
		case sign == 0:
			*on = !*on
			st.Both = st.Both && !*on
		case *on:
			*way = sign
			md.persistStep()
			md.stepFrom(ctx, a.Node, a.K, sign)
			return
		default:
			md.stepFrom(ctx, a.Node, a.K, sign)
			return
		}
		if *way == 0 {
			*way = 1
		}
		md.persistStep()
	case Builder.OpAllBoth:
		if sign == 0 {
			sign = a.Sign
		}
		md.stepFrom(ctx, a.Node, Links.KBoth, sign)
	case Builder.OpAllLocks:
		md.UI.Builder.NodeStep(a.Node).ToggleBoth()
		md.persistStep()
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
