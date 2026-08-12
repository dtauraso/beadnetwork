package dispatch

import (
	"context"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/gesture"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
)

func (md *MoveDispatch) HandleRawInput(ctx context.Context, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
	gesture.HandleRawInput(gesture.Deps{MR: &md.MR, UI: &md.UI, LQ: &md.LQ, RT: &md.RT, Ctx: ctx}, ev, slotReg, tr)
}
