package Dispatch

import (
	"context"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
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
		if sign != 0 {
			md.sendLink(ctx, a.Node, Links.Press{To: a.To, K: a.K, Sign: sign})
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
