package scenerun

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Input/Drag"
	"github.com/dtauraso/wirefold/Categories/Input/Gesture"
)

func (md *MoveDispatch) HandleRawInput(ctx context.Context, ev Drag.RawInputMsg) {
	Gesture.HandleRawInput(Gesture.Deps{MR: &md.MR, UI: &md.UI, Mover: &md.Mover, RT: &md.RT, Ctx: ctx}, ev)
}
