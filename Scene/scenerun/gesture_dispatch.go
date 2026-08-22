package scenerun

import (
	"context"

	"github.com/dtauraso/wirefold/Input/Drag"
	"github.com/dtauraso/wirefold/Input/Gesture"
	beadanimation "github.com/dtauraso/wirefold/Node/BeadAnimation"
)

func (md *MoveDispatch) HandleRawInput(ctx context.Context, ev Drag.RawInputMsg, slotReg beadanimation.SlotRegistry) {
	Gesture.HandleRawInput(Gesture.Deps{MR: &md.MR, UI: &md.UI, Mover: &md.Mover, RT: &md.RT, Ctx: ctx}, ev, slotReg)
}
