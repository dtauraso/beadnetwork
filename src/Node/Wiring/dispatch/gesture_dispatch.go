package dispatch

import (
	"context"

	"github.com/dtauraso/wirefold/src/Node/Wiring/gesture"
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
)

func (md *MoveDispatch) HandleRawInput(ctx context.Context, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry) {
	gesture.HandleRawInput(gesture.Deps{MR: &md.MR, UI: &md.UI, Mover: &md.Mover, RT: &md.RT, Ctx: ctx}, ev, slotReg)
}
