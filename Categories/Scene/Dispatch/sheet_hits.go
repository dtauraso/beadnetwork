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
	case Sheet.PartUp:
		runSheetAct(ctx, md, h.Cell.ArrowAct(), 1, nil)
	case Sheet.PartDown:
		runSheetAct(ctx, md, h.Cell.ArrowAct(), -1, nil)
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
		if typed != nil {
			v = *typed
		}
		if a.K == Links.KR && (v < 0 || v+l.DeltaR < 0) {
			md.UI.Builder.Refused = fmt.Sprintf("%s→%s r %v: neither r nor r + Δr can be negative, so r stays at %v", a.Node, a.To, v, l.Start.R)
			return
		}
		md.sendLink(ctx, a.To, Links.SetStart{From: a.Node, K: a.K, V: v})
	case Builder.OpLinkPress:
		if sign == 0 {
			sign = a.Sign
		}
		if sign != 0 {
			md.sendLink(ctx, a.To, Links.Press{From: a.Node, K: a.K, Sign: sign})
		}
	case Builder.OpUnfold:
		m := &md.UI.Builder
		m.ToggleUnfold(a.K)
		if m.UnfoldDir != "" {
			if err := Builder.WriteUnfold(m.UnfoldDir, m.Unfold); err != nil {
				View.LogPersistErr("sheet_unfold_persist", m.UnfoldDir, err)
			}
		}
	case Builder.OpPick:
		md.sendLink(ctx, a.To, Links.PickReq{From: a.Node})
	case Builder.OpUnpick:
		md.sendLink(ctx, a.To, Links.Unpick{From: a.Node})
	case Builder.OpIncrement:
		md.setIncrement(ctx, a, sign, typed)
	case Builder.OpLocking:
		if l, ok := m.Link(a.Node, a.To); ok {
			md.sendLink(ctx, a.To, Links.SetLocking{From: a.Node, On: !l.Locking})
		}
	case Builder.OpItem, Builder.OpMethod, Builder.OpCard:
		setCards(md, a)
	case Builder.OpAllStep:
		if sign == 0 {
			sign = a.Sign
		}
		allStep(ctx, md, a.K, sign)
	case Builder.OpAllLock:
		lock := allLockOf(md, a.K)
		if *lock == sign {
			sign = 0
		}
		*lock = sign
		md.persistStep()
	case Builder.OpAllBoth:
		if sign == 0 {
			sign = a.Sign
		}
		md.stepAll(ctx, Links.KBoth, sign)
	case Builder.OpAllLocking:
		st := &md.UI.Builder.Step
		if !m.RowsLocking() {
			st.Locking = !st.Locking
			md.persistStep()
		}
	}
}

func allLockOf(md *MoveDispatch, k string) *int {
	st := &md.UI.Builder.Step
	switch k {
	case Links.KTheta:
		return &st.LockAll.Theta
	case Links.KR:
		return &st.LockAll.R
	}
	return &st.LockAll.Phi
}

func allStep(ctx context.Context, md *MoveDispatch, k string, sign int) {
	st := &md.UI.Builder.Step
	lock := allLockOf(md, k)
	switch {
	case *lock == sign:
		*lock = 0
	case st.Locking:
		*lock = sign
	default:
		md.stepAll(ctx, k, sign)
		return
	}
	md.persistStep()
}

func (md *MoveDispatch) stepAll(ctx context.Context, k string, sign int) {
	st := &md.UI.Builder.Step
	st.Round++
	h := clock.HumanTick(md.UI.Builder.TickMs)
	for _, id := range md.UI.Builder.Nodes {
		md.sendLink(ctx, id, Links.StepAll{Round: st.Round, H: h, K: k, Sign: sign, LockAll: st.LockAll})
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

func (md *MoveDispatch) persistStep() {
	if dir := md.UI.Builder.StepDir; dir != "" {
		if err := Builder.WriteStep(dir, md.UI.Builder.Step); err != nil {
			View.LogPersistErr("sheet_step_persist", dir, err)
		}
	}
}

func DrainLinkReports(md *MoveDispatch) {
	if md == nil || md.Links == nil {
		return
	}
	m := &md.UI.Builder
	rs := md.Links.TakeReports()
	for _, r := range rs {
		m.Take(r)
	}
	before := m.TimeKey()
	for _, t := range md.Links.TakeTimings() {
		m.TakeTiming(t)
	}
	m.Now = clock.HumanTick(m.TickMs)
	if len(rs) > 0 || m.TimeKey() != before {
		md.UI.EmitViewFrame(nil)
	}
}
