package stdinreader

import (
	"context"
	"github.com/dtauraso/wirefold/Slider"

	"github.com/dtauraso/wirefold/nodes/rowevent"

	T "github.com/dtauraso/wirefold/Trace"
	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

func HandleRawInputMsg(ctx context.Context, msg inputcodec.StdinMsg, slotReg inputcodec.SlotRegistry, md *dispatch.MoveDispatch, tr *T.Trace) {
	if md != nil && msg.Event != nil {
		md.HandleRawInput(ctx, *msg.Event, slotReg, tr)
	}
}

func HandleSaveMsg(md *dispatch.MoveDispatch) {
	if md == nil {
		return
	}
	md.Persist.Overlays().Schedule(md.UI.OV)

	md.Persist.Panels().Schedule(md.UI.PN)

	md.Persist.Sphere().Schedule(md.UI.SceneSphere)
}

// EDIT_OPS_START
var editOps = map[string]func(context.Context, inputcodec.StdinMsg, *dispatch.MoveDispatch, *T.Trace, Slider.Sinks){
	"update": applyUpdate,
}

// EDIT_OPS_END

func ApplyEdit(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace, speedSinks Slider.Sinks) {
	if h, ok := editOps[msg.Op]; ok {
		h(ctx, msg, md, tr, speedSinks)
	}
}

// EDIT_UPDATE_KINDS_START
var updateKindHandlers = map[string]func(context.Context, inputcodec.StdinMsg, *dispatch.MoveDispatch, *T.Trace, Slider.Sinks){
	"clock":         applyUpdateClock,
	"overlays":      applyUpdateOverlays,
	"distanceGroup": applyUpdateDistanceGroup,
	"scene":         applyUpdateScene,
	"tiltVector":    applyUpdateTiltVector,
	"panels":        applyUpdatePanels,
	"node":          applyUpdateNode,
	"edge":          applyUpdateEdge,
}

// EDIT_UPDATE_KINDS_END

func applyUpdate(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace, speedSinks Slider.Sinks) {
	if h, ok := updateKindHandlers[msg.Kind]; ok {
		h(ctx, msg, md, tr, speedSinks)
	}
}

var clockAttrHandlers = map[string]func(msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, speedSinks Slider.Sinks){
	"speed": func(msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, speedSinks Slider.Sinks) {

		divisor := int64(1)
		if md != nil {
			divisor = int64(md.UI.ClockDivisor)
		}

		Slider.Broadcast(speedSinks, int64(msg.Num), divisor)

		userSpeed := float64(msg.Num) / Slider.NumScale
		if md == nil {
			return
		}

		md.UI.Speed = userSpeed
		md.Persist.Speed().Schedule(userSpeed)
		md.UI.EmitViewFrame(nil)
	},
}

var panelAttrHandlers = map[string]func(msg inputcodec.StdinMsg, md *dispatch.MoveDispatch){
	"toggle": func(msg inputcodec.StdinMsg, md *dispatch.MoveDispatch) {
		if fn, ok := viewstate.PanelToggles[msg.Flag]; ok {
			fn(&md.UI.PN)
		}
	},
}

var overlayAttrHandlers = map[string]func(msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace){
	"toggle": func(msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace) {

		if fn, ok := viewstate.OverlayToggles[msg.Flag]; ok {
			fn(&md.UI.OV, tr)

			if scope, ok := viewstate.OverlayFlagBreadcrumbScope[msg.Flag]; ok {
				md.UI.EmitBreadcrumb(rowevent.RowEvent{Label: T.BreadcrumbPoleToggleGo, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1, Value: int32(boolU8(viewstate.OverlayFlagValue[msg.Flag](&md.UI.OV))), Text: scope})
			}

			if kind, ok := viewstate.OverlayFlagTraceKind[msg.Flag]; ok {
				md.UI.EmitViewFrame([]rowevent.RowEvent{{Kind: kind, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1}})
			}
		}
	},
}
