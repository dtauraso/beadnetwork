package Dispatch

import (
	"context"

	"github.com/dtauraso/wirefold/src/Input/Codec"
	"github.com/dtauraso/wirefold/src/Input/Gesture"
)

func (md *MoveDispatch) HandleRawInput(ctx context.Context, ev Codec.RawInputMsg, slotReg Codec.SlotRegistry) {
	Gesture.HandleRawInput(Gesture.Deps{MR: &md.MR, UI: &md.UI, Mover: &md.Mover, RT: &md.RT, Ctx: ctx}, ev, slotReg)
}
