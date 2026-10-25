package Dispatch

import (
	"context"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
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
		runSheetAct(ctx, md, h.Cell.Act, 1, nil)
	case Sheet.PartDown:
		runSheetAct(ctx, md, h.Cell.Act, -1, nil)
	}
	md.UI.EmitViewFrame(nil)
}

func applySheetKey(ctx context.Context, md *MoveDispatch, key string) {
	if act, v, ok := Sheet.Key(&md.UI.Sheet.Edit, key); ok {
		runSheetAct(ctx, md, act, 0, &v)
	}
	md.UI.EmitViewFrame(nil)
}

func (md *MoveDispatch) sendLink(ctx context.Context, to string, msg any) {
	if md.Links == nil {
		return
	}
	md.Links.Send(ctx, "", Links.Envelope{To: to, Msg: msg})
}

func runSheetAct(ctx context.Context, md *MoveDispatch, a Sheet.Act, sign int, typed *float64) {
	m := md.UI.Sheet.Model
	md.UI.Sheet.Model.Refused = ""
	if runSheetStructure(ctx, md, a, sign, typed) || runSheetSequence(md, a) {
		return
	}
	switch a.Op {
	case Sheet.OpLinkStart:
		l, ok := m.Link(a.Node, a.To)
		if !ok {
			return
		}
		v := float64(sign) + map[string]float64{Links.KPhi: l.Start.Phi, Links.KTheta: l.Start.Theta, Links.KR: l.Start.R}[a.K]
		if typed != nil {
			v = *typed
		}
		md.sendLink(ctx, a.Node, Links.SetStart{To: a.To, K: a.K, V: v})
	case Sheet.OpLinkPress:
		if sign == 0 {
			sign = a.Sign
		}
		if sign != 0 {
			md.sendLink(ctx, a.Node, Links.Press{To: a.To, K: a.K, Sign: sign})
		}
	case Sheet.OpPick:
		md.sendLink(ctx, a.To, Links.PickReq{From: a.Node})
	case Sheet.OpUnpick:
		md.sendLink(ctx, a.Node, Links.Unpick{To: a.To})
	case Sheet.OpIncrement:
		if typed != nil {
			md.sendLink(ctx, a.Node, Links.SetIncrement{To: a.To, K: a.K, V: *typed})
		}
	case Sheet.OpLocking:
		if l, ok := m.Link(a.Node, a.To); ok {
			md.sendLink(ctx, a.Node, Links.SetLocking{To: a.To, On: !l.Locking})
		}
	case Sheet.OpAllStep:
		allStep(ctx, md, a.K, a.Sign)
	case Sheet.OpAllBoth:
		md.stepAll(ctx, Links.KBoth, a.Sign)
	case Sheet.OpAllLocking:
		st := &md.UI.Sheet.Model.Step
		if !m.RowsLocking() {
			st.Locking = !st.Locking
			md.persistStep()
		}
	}
}

func allStep(ctx context.Context, md *MoveDispatch, k string, sign int) {
	st := &md.UI.Sheet.Model.Step
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
	st := &md.UI.Sheet.Model.Step
	st.Round++
	for _, id := range md.UI.Sheet.Model.Nodes {
		md.sendLink(ctx, id, Links.StepAll{Round: st.Round, K: k, Sign: sign, LockAll: st.LockAll})
	}
}

func (md *MoveDispatch) persistStep() {
	if dir := md.UI.Sheet.Model.StepDir; dir != "" {
		if err := Sheet.WriteStep(dir, md.UI.Sheet.Model.Step); err != nil {
			View.LogPersistErr("sheet_step_persist", dir, err)
		}
	}
}

func DrainLinkReports(md *MoveDispatch) {
	if md == nil || md.Links == nil {
		return
	}
	rs := md.Links.TakeReports()
	for _, r := range rs {
		md.UI.Sheet.Model.Take(r)
	}
	if len(rs) > 0 {
		md.UI.EmitViewFrame(nil)
	}
}
