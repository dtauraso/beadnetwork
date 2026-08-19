package dispatch

import (
	"context"

	"github.com/dtauraso/wirefold/src/Node/Wiring/gesture"
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
	T "github.com/dtauraso/wirefold/src/Trace"
)

func (md *MoveDispatch) HandleRawInput(ctx context.Context, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry, tr *T.Trace) {
	gesture.HandleRawInput(gesture.Deps{MR: &md.MR, UI: &md.UI, Mover: &md.Mover, RT: &md.RT, Ctx: ctx}, ev, slotReg, tr)
}
