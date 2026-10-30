package Dispatch

import (
	"context"
	"fmt"
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
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
		runSheetAct(ctx, md, h.Cell.Act, 1, nil)
	case Sheet.PartDown:
		runSheetAct(ctx, md, h.Cell.Act, -1, nil)
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
		by := 1.0
		if a.K == Links.KR {
			by = l.IncR
		}
		v := float64(sign)*by + map[string]float64{Links.KPhi: l.Start.Phi, Links.KTheta: l.Start.Theta, Links.KR: l.Start.R}[a.K]
		if typed != nil {
			v = *typed
		}
		if a.K == Links.KR && v < 0 {
			md.UI.Builder.Refused = fmt.Sprintf("%s→%s r %v: a length cannot be negative, so it stays at %v", a.Node, a.To, v, l.Start.R)
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
		md.UI.Builder.ToggleUnfold(a.K)
	case Builder.OpPick:
		md.sendLink(ctx, a.To, Links.PickReq{From: a.Node})
	case Builder.OpUnpick:
		md.sendLink(ctx, a.To, Links.Unpick{From: a.Node})
	case Builder.OpIncrement:
		if typed != nil {
			md.setIncrement(ctx, a, *typed)
		}
	case Builder.OpLocking:
		if l, ok := m.Link(a.Node, a.To); ok {
			md.sendLink(ctx, a.To, Links.SetLocking{From: a.Node, On: !l.Locking})
		}
	case Builder.OpItem, Builder.OpMethod, Builder.OpCard:
		setCards(md, a)
	case Builder.OpAllStep:
		allStep(ctx, md, a.K, a.Sign)
	case Builder.OpAllBoth:
		md.stepAll(ctx, Links.KBoth, a.Sign)
	case Builder.OpAllLocking:
		st := &md.UI.Builder.Step
		if !m.RowsLocking() {
			st.Locking = !st.Locking
			md.persistStep()
		}
	}
}

func allStep(ctx context.Context, md *MoveDispatch, k string, sign int) {
	st := &md.UI.Builder.Step
	lock := &st.LockAll.Phi
	if k == Links.KTheta {
		lock = &st.LockAll.Theta
	}
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

func setTickMs(ctx context.Context, md *MoveDispatch, sign int, typed *float64) {
	m := &md.UI.Builder
	v := m.TickMs + sign
	if typed != nil {
		if !whole(*typed) || *typed < 1 {
			m.Refused = fmt.Sprintf("Δt %v: not a whole number of milliseconds of at least 1, so it stays at %d", *typed, m.TickMs)
			return
		}
		v = int(math.Round(*typed))
	}
	if v < 1 {
		return
	}
	m.TickMs = v
	if root := md.Scenes.TreeRoot; root != "" {
		if err := Sheet.WriteScalar(Scenes.TickMsFilePath(root), int32(v)); err != nil {
			m.Refused = fmt.Sprintf("could not write Δt: %v", err)
		}
	}
	md.sendLink(ctx, Links.All, Links.SetTickMs{V: v})
}
