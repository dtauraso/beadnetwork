package Dispatch

import (
	"context"
	"fmt"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Scene/Drag"

	"github.com/dtauraso/wirefold/Categories/Scene/structuraledit"
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
				nodeID, _ := md.RT.LookupNodeRow(int(row))
				structuraledit.DeleteNode(&md.Scenes, &md.UI, nodeID, int(row))
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

func ApplyEdit(ctx context.Context, op string, entity, attr byte, payload []byte, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
	if h, ok := editOps[op]; ok {
		h(ctx, entity, attr, payload, md, speedSinks)
	}
}

func init() {
	// Checked against the OWNER TABLE itself, not against a parallel list of
	// names: a list could agree with the wire and still name an entity no owner
	// handles. Building it on a zero composer is safe — the closures are never
	// called here, only counted.
	named := (&MoveDispatch{}).updateOwners(SliderPanel.Sinks{})
	var missing []string
	for _, entity := range Drag.UpdateKinds {
		if _, ok := named[entity]; !ok {
			missing = append(missing, entity)
		}
	}
	if len(missing) > 0 {
		panic(fmt.Sprintf(
			"dispatch: Drag.UpdateKinds names %v but no owner is named for them. The wire can "+
				"carry an edit for each name here, so every one of those edits would decode cleanly and "+
				"then be dropped — the click does nothing and nothing reports why.",
			missing))
	}
}

func applyUpdate(ctx context.Context, entity, attr byte, payload []byte, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
	if md == nil || int(entity) >= len(Drag.UpdateKinds) {
		return
	}
	if h, ok := md.updateOwners(speedSinks)[Drag.UpdateKinds[entity]]; ok {
		h(ctx, attr, payload)
	}
}
