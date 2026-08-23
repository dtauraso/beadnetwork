package scenerun

import (
	"context"

	"github.com/dtauraso/wirefold/Categories/Scene/Drag"
	"github.com/dtauraso/wirefold/Categories/Scene/Gesture"
)

func (md *MoveDispatch) HandleRawInput(ctx context.Context, ev Drag.RawInputMsg) {
	Gesture.HandleRawInput(Gesture.Deps{MR: md.gestureMovers(), UI: &md.UI, Mover: &md.Mover, RT: &md.RT, Ctx: ctx}, ev)
}
