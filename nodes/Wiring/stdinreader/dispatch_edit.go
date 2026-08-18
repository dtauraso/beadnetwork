package stdinreader

import (
	"context"
	"github.com/dtauraso/wirefold/tools/topology-vscode/OverlaysDropdown"
	"github.com/dtauraso/wirefold/tools/topology-vscode/SliderPanel"

	"github.com/dtauraso/wirefold/nodes/rowevent"

	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/nodes/Wiring/speedpanel"
	T "github.com/dtauraso/wirefold/tools/topology-vscode/Trace"
)

func HandleRawInputMsg(ctx context.Context, msg inputcodec.StdinMsg, slotReg inputcodec.SlotRegistry, md *dispatch.MoveDispatch, tr *T.Trace, speedSinks SliderPanel.Sinks) {
	if md == nil || msg.Event == nil {
		return
	}
	if msg.Event.Kind == "pointerdown" {
		if i := speedpanel.Hit(msg.Event.X, msg.Event.Y); i >= 0 {
			setClockSpeed(md, speedSinks, speedpanel.Settings[i].Speed)
			return
		}
	}
	md.HandleRawInput(ctx, *msg.Event, slotReg, tr)
}

func setClockSpeed(md *dispatch.MoveDispatch, speedSinks SliderPanel.Sinks, speed float64) {
	divisor := int64(md.UI.ClockDivisor)
	SliderPanel.Broadcast(speedSinks, int64(speed*SliderPanel.NumScale), divisor)
	md.UI.Speed = speed
	md.Persist.Speed().Schedule(speed)
	md.UI.EmitViewFrame(nil)
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
var editOps = map[string]func(context.Context, inputcodec.StdinMsg, *dispatch.MoveDispatch, *T.Trace, SliderPanel.Sinks){
	"update": applyUpdate,
}

// EDIT_OPS_END

func ApplyEdit(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace, speedSinks SliderPanel.Sinks) {
	if h, ok := editOps[msg.Op]; ok {
		h(ctx, msg, md, tr, speedSinks)
	}
}

// EDIT_UPDATE_KINDS_START
var updateKindHandlers = map[string]func(context.Context, inputcodec.StdinMsg, *dispatch.MoveDispatch, *T.Trace, SliderPanel.Sinks){
	"clock":      applyUpdateClock,
	"overlays":   applyUpdateOverlays,
	"scene":      applyUpdateScene,
	"tiltVector": applyUpdateTiltVector,
	"panels":     applyUpdatePanels,
	"node":       applyUpdateNode,
	"edge":       applyUpdateEdge,
}

// EDIT_UPDATE_KINDS_END

func applyUpdate(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace, speedSinks SliderPanel.Sinks) {
	if h, ok := updateKindHandlers[msg.Kind]; ok {
		h(ctx, msg, md, tr, speedSinks)
	}
}

var clockAttrHandlers = map[string]func(msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, speedSinks SliderPanel.Sinks){
	"speed": func(msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, speedSinks SliderPanel.Sinks) {

		divisor := int64(1)
		if md != nil {
			divisor = int64(md.UI.ClockDivisor)
		}

		SliderPanel.Broadcast(speedSinks, int64(msg.Num), divisor)

		userSpeed := float64(msg.Num) / SliderPanel.NumScale
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
		if fn, ok := OverlaysDropdown.PanelToggles[msg.Flag]; ok {
			fn(&md.UI.PN)
		}
	},
}

var overlayAttrHandlers = map[string]func(msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace){
	"toggle": func(msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, tr *T.Trace) {

		if fn, ok := OverlaysDropdown.OverlayToggles[msg.Flag]; ok {
			fn(&md.UI.OV, tr)

			if msg.Flag == "ruleChannels" {
				md.Inboxes.BroadcastChannelVectorsOn(md.UI.OV.RuleChannelsVisible)
			}

			if scope, ok := OverlaysDropdown.OverlayFlagBreadcrumbScope[msg.Flag]; ok {
				md.UI.EmitBreadcrumb(rowevent.RowEvent{Label: T.BreadcrumbPoleToggleGo, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1, Value: int32(boolU8(OverlaysDropdown.OverlayFlagValue[msg.Flag](&md.UI.OV))), Text: scope})
			}

			md.UI.EmitViewFrame(nil)
		}
	},
}
