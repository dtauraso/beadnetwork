package stdinreader

import (
	"context"

	"github.com/dtauraso/wirefold/tools/topology-vscode/OverlaysDropdown"
	"github.com/dtauraso/wirefold/tools/topology-vscode/SliderPanel"

	"github.com/dtauraso/wirefold/nodes/Wiring/dispatch"
	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/tools/topology-vscode/NodesDropdown"
	T "github.com/dtauraso/wirefold/tools/topology-vscode/Trace"
)

func HandleRawInputMsg(ctx context.Context, msg inputcodec.StdinMsg, slotReg inputcodec.SlotRegistry, md *dispatch.MoveDispatch, tr *T.Trace, speedSinks SliderPanel.Sinks) {
	if md == nil || msg.Event == nil {
		return
	}
	if msg.Event.RectWidth > 0 && msg.Event.RectHeight > 0 {
		md.UI.ViewW = msg.Event.RectWidth
		md.UI.ViewH = msg.Event.RectHeight
	}
	if msg.Event.Kind == "key" {
		applyRuleKey(ctx, md, msg.Event.Key)
		return
	}
	if msg.Event.Kind == "delete" {
		if md.UI.SceneEditable && md.UI.Sel.Selected != "" {
			if row, ok := md.UI.NodeRowFor(md.UI.Sel.Selected); ok {
				NodesDropdown.DeleteNode(&md.Scenes, &md.UI, &md.RT, int(row), tr)
			}
		}
		return
	}
	if msg.Event.Kind == "pointerup" && md.UI.PlacingPending {
		md.UI.PlacingPending = false
		placeNodeAt(md, msg.Event, tr)
		return
	}
	if msg.Event.Kind == "pointermove" {
		if t := panelPointerTarget(md, msg.Event.X, msg.Event.Y); t != md.UI.Pointer {
			md.UI.Pointer = t
			md.UI.EmitViewFrame(nil)
		}
	}
	if msg.Event.Kind == "wheel" && panelTookWheel(*msg.Event, md) {
		return
	}
	if msg.Event.Kind == "pointerdown" && panelTookPointerDown(ctx, *msg.Event, md, tr, speedSinks) {
		return
	}
	md.HandleRawInput(ctx, *msg.Event, slotReg, tr)
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
		toggleOverlayFlag(md, tr, msg.Flag)
		md.UI.EmitViewFrame(nil)
	},
}
