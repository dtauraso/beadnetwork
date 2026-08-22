package Dispatch

import (
	"context"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/src/Input/Drag"
	"github.com/dtauraso/wirefold/src/Input/Stdin"

	beadanimation "github.com/dtauraso/wirefold/src/Node/BeadAnimation"
	"github.com/dtauraso/wirefold/src/Node/nodecrud"
)

func HandleRawInputMsg(ctx context.Context, ev Drag.RawInputMsg, slotReg beadanimation.SlotRegistry, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
	if md == nil {
		return
	}
	if ev.RectWidth > 0 && ev.RectHeight > 0 {
		md.UI.ViewW = ev.RectWidth
		md.UI.ViewH = ev.RectHeight
	}
	if ev.Kind == "key" {
		applyRuleKey(ctx, md, ev.Key)
		return
	}
	if ev.Kind == "delete" {
		if md.UI.SceneEditable && md.UI.Sel.Selected != "" {
			if row, ok := md.UI.NodeRowFor(md.UI.Sel.Selected); ok {
				nodecrud.DeleteNode(&md.Scenes, &md.UI, &md.RT, int(row))
			}
		}
		return
	}
	if ev.Kind == "pointerup" && md.UI.PlacingPending {
		md.UI.PlacingPending = false
		placeNodeAt(md, &ev)
		return
	}
	if ev.Kind == "pointermove" {
		if t := md.UI.PointerTargetAt(ev.X, ev.Y); t != md.UI.Pointer {
			md.UI.Pointer = t
			md.UI.EmitViewFrame(nil)
		}
	}
	if ev.Kind == "wheel" && md.UI.TakeWheel(ev.X, ev.Y, ev.DeltaY) {
		return
	}
	if ev.Kind == "pointerdown" && panelTookPointerDown(ctx, ev, md, speedSinks) {
		return
	}
	md.HandleRawInput(ctx, ev, slotReg)
}

func HandleSaveMsg(md *MoveDispatch) {
	if md == nil {
		return
	}
	md.Persist.Overlays().Schedule(md.UI.OV)

	md.Persist.Panels().Schedule(md.UI.PN)

	md.Persist.Sphere().Schedule(md.UI.SceneSphere)
}

// EDIT_OPS_START
var editOps = map[string]func(context.Context, Stdin.StdinMsg, *MoveDispatch, SliderPanel.Sinks){
	"update": applyUpdate,
}

// EDIT_OPS_END

func ApplyEdit(ctx context.Context, msg Stdin.StdinMsg, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
	if h, ok := editOps[msg.Op]; ok {
		h(ctx, msg, md, speedSinks)
	}
}

// EDIT_UPDATE_KINDS_START
// Every entity the wire can name has an entry, and the entry is where the composer is
// UNPACKED. A handler below the line takes the fields it uses; the one-line adapter here is
// what turns MoveDispatch into those fields, and is what gets deleted when the handler moves
// to its own concern and registers itself.
var updateKindHandlers = map[string]func(context.Context, Stdin.StdinMsg, *MoveDispatch, SliderPanel.Sinks){
	"clock":      applyUpdateClock,
	"scene":      applyUpdateScene,
	"tiltVector": applyUpdateTiltVector,
	"node":       applyUpdateNode,
	"edge":       applyUpdateEdge,

	"overlays": applyUpdateOverlays,
	"panels":   applyUpdatePanels,
}

// EDIT_UPDATE_KINDS_END

func applyUpdate(ctx context.Context, msg Stdin.StdinMsg, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
	if md == nil {
		return
	}
	if h, ok := updateKindHandlers[msg.Kind]; ok {
		h(ctx, msg, md, speedSinks)
	}
}
