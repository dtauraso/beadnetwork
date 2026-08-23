package scenerun

import (
	"context"
	"fmt"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Input/Drag"

	"github.com/dtauraso/wirefold/Categories/Node/nodecrud"
)

func HandleRawInputMsg(ctx context.Context, ev Drag.RawInputMsg, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
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
	md.HandleRawInput(ctx, ev)
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
var editOps = map[string]func(context.Context, byte, byte, []byte, *MoveDispatch, SliderPanel.Sinks){
	"update": applyUpdate,
}

// EDIT_OPS_END

const KindEditUpdate = 22

var UpdateKinds = []string{
	"overlays",
	"clock",
	"scene",
	"tiltVector",
	"panels",
	"node",
	"edge",
}

func ApplyEdit(ctx context.Context, op string, entity, attr byte, payload []byte, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
	if h, ok := editOps[op]; ok {
		h(ctx, entity, attr, payload, md, speedSinks)
	}
}

// EDIT_UPDATE_KINDS_START
// Every entity the wire can name has an entry, and the entry is where the composer is
// UNPACKED. A handler below the line takes the fields it uses; the one-line adapter here is
// what turns MoveDispatch into those fields, and is what gets deleted when the handler moves
// to its own concern and registers itself.
var updateKindHandlers = map[string]func(context.Context, byte, []byte, *MoveDispatch, SliderPanel.Sinks){
	"clock":      applyUpdateClock,
	"scene":      applyUpdateScene,
	"tiltVector": applyUpdateTiltVector,
	"node":       applyUpdateNode,
	"edge":       applyUpdateEdge,

	"overlays": applyUpdateOverlays,
	"panels":   applyUpdatePanels,
}

// EDIT_UPDATE_KINDS_END

func init() {
	var missing []string
	for _, entity := range UpdateKinds {
		if _, ok := updateKindHandlers[entity]; !ok {
			missing = append(missing, entity)
		}
	}
	if len(missing) > 0 {
		panic(fmt.Sprintf(
			"scenerun: UpdateKinds names %v but updateKindHandlers has no entry for them. The wire can "+
				"carry an edit for each name here, so every one of those edits would decode cleanly and "+
				"then be dropped — the click does nothing and nothing reports why.",
			missing))
	}
}

func applyUpdate(ctx context.Context, entity, attr byte, payload []byte, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
	if md == nil || int(entity) >= len(UpdateKinds) {
		return
	}
	if h, ok := updateKindHandlers[UpdateKinds[entity]]; ok {
		h(ctx, attr, payload, md, speedSinks)
	}
}
