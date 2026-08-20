package dispatch

import (
	"context"

	"github.com/dtauraso/wirefold/src/Input/gesture"
	"github.com/dtauraso/wirefold/src/Input/inputcodec"
)

func (md *MoveDispatch) HandleRawInput(ctx context.Context, ev inputcodec.RawInputMsg, slotReg inputcodec.SlotRegistry) {
	gesture.HandleRawInput(gesture.Deps{MR: &md.MR, UI: &md.UI, Mover: &md.Mover, RT: &md.RT, Ctx: ctx}, ev, slotReg)
}
