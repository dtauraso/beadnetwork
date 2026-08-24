package Dispatch

import (
	"context"
	"fmt"

	Chrome "github.com/dtauraso/beadnetwork/Categories/Chrome"

	NodeKind "github.com/dtauraso/beadnetwork/Categories/Node"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	edge "github.com/dtauraso/beadnetwork/Categories/Node/Edge"
	Flags "github.com/dtauraso/beadnetwork/Categories/Scene/View/Flags"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Drag"
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
		if sel, any := md.UI.SelectedNode(); md.UI.SceneEditable && any {
			if row, ok := md.UI.NodeRowFor(sel); ok {
				nodeID, _ := md.RT.LookupNodeRow(int(row))
				DeleteNode(&md.Scenes, &md.UI, nodeID, int(row))
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
		if t := Chrome.TargetAt(md.UI.PanelLayout(), ev.X, ev.Y); t != md.UI.Pointer {
			md.UI.Pointer = t
			md.UI.EmitViewFrame(nil)
		}
	}
	if ev.Kind == "wheel" && Chrome.TakeWheel(md.UI.PanelLayout(), &md.UI.OverlaysPill.Scroll, &md.UI.Rules.Scroll, ev.X, ev.Y, ev.DeltaY, md.redraw) {
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
	md.UI.PersistOverlays(md.UI.OV)

	md.UI.PersistPanels(md.UI.PN)

	md.UI.PersistSphere(md.UI.SceneSphere)
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

// EDIT_UPDATE_KINDS_START
func (md *MoveDispatch) updateOwners(speedSinks SliderPanel.Sinks) map[string]func(context.Context, byte, []byte) {
	return map[string]func(context.Context, byte, []byte){
		"node": func(ctx context.Context, attr byte, payload []byte) {
			NodeKind.ApplyUpdate(ctx, attr, payload, &md.Rules)
		},
		"edge": func(ctx context.Context, attr byte, payload []byte) {
			edge.ApplyUpdate(ctx, attr, payload, md.Rules.TogglesByEdgeRow)
		},
		"overlays": func(_ context.Context, attr byte, payload []byte) {
			Flags.ApplyUpdate(attr, payload, &md.UI.OV, &md.ChannelVectorsOn, &md.UI, md.persistOverlays)
		},
		"panels": func(_ context.Context, attr byte, payload []byte) {
			Panel.ApplyUpdate(attr, payload, &md.UI.PN, md.persistPanels, md.redraw)
		},
		"clock": func(_ context.Context, attr byte, payload []byte) {
			SliderPanel.ApplyUpdate(attr, payload, md.speedState(), speedSinks, md.persistSpeed, md.redraw)
		},
		"tiltVector": func(ctx context.Context, attr byte, payload []byte) {
			applyUpdateTiltVector(ctx, attr, payload, md, speedSinks)
		},
		"scene": func(ctx context.Context, attr byte, payload []byte) {
			applyUpdateScene(ctx, attr, payload, md, speedSinks)
		},
	}
}

// EDIT_UPDATE_KINDS_END
