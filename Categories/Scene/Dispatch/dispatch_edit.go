package Dispatch

import (
	"context"
	"fmt"

	Chrome "github.com/dtauraso/beadnetwork/Categories/Chrome"

	Panels "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Drag"
	Flags "github.com/dtauraso/beadnetwork/Categories/Scene/View/Flags"
)

func HandleRawInputMsg(ctx context.Context, ev Drag.RawInputMsg, md *MoveDispatch, speedSinks SliderPanel.Sinks) {
	if md == nil {
		return
	}
	if ev.Kind == "key" {
		if md.UI.Sheet.Edit.Active {
			treeFinding := md.UI.Sheet.Edit.Act.Op == Builder.OpTreeFind
			applySheetKey(ctx, md, ev.Key)
			if treeFinding {
				md.UI.Builder.FindStep(md.UI.Sheet.Edit, ev.Key)
				md.UI.EmitViewFrame(nil)
			}
			return
		}
		return
	}
	if ev.Kind == "delete" {
		if sel, any := md.UI.SelectedNode(); md.UI.SceneEditable && any {
			if row, ok := md.UI.NodeRowFor(sel); ok {
				nodeID, _ := md.RT.LookupNodeRow(int(row))
				if md.UI.Builder.On() {
					removeLinkNode(md, nodeID)
					return
				}
				DeleteNode(&md.Scenes, &md.UI, nodeID, int(row))
			}
		}
		return
	}
	if ev.Kind == "pointermove" && md.UI.Sheet.Resize(md.UI.ViewH, ev.Y) {
		md.UI.EmitViewFrame(nil)
		return
	}
	if ev.Kind == "pointerup" && md.UI.Sheet.EndResize() {
		return
	}
	if ev.Kind == "pointermove" && md.UI.Builder.SplitTo(ev.X) {
		md.UI.EmitViewFrame(nil)
		return
	}
	if ev.Kind == "pointerup" && md.UI.Builder.EndSplit() {
		return
	}
	if ev.Kind == "pointermove" && md.UI.Sheet.DragTab(md.UI.PanelLayout().Sheet, ev.X) {
		md.UI.EmitViewFrame(nil)
		return
	}
	if ev.Kind == "pointerup" && md.UI.Sheet.EndTabDrag() {
		return
	}
	if ev.Kind == "pointerup" && md.UI.PlacingPending {
		md.UI.PlacingPending = false
		if Chrome.TargetAt(md.UI.PanelLayout(), ev.X, ev.Y).Kind == Panels.PointerNothing {
			placeNodeAt(md, &ev)
		}
		return
	}
	if ev.Kind == "pointermove" {
		t := Chrome.TargetAt(md.UI.PanelLayout(), ev.X, ev.Y)
		hot := Sheet.Act{}
		if h := md.UI.PanelLayout().Sheet.Hit(ev.X, ev.Y); h.Part == Sheet.PartBody {
			hot = h.Cell.Act
		} else if h.Part == Sheet.PartNone && ev.Hit.Kind == "edge" && ev.Hit.EdgeRow >= 0 && int(ev.Hit.EdgeRow) < len(md.RT.EdgeEndpointRowTable) {
			e := md.RT.EdgeEndpointRowTable[ev.Hit.EdgeRow]
			hot = md.UI.Builder.EdgeHot(e.SrcNode, e.DstNode)
		}
		if hovered := md.UI.Builder.Hover(hot); t != md.UI.Pointer || hovered {
			md.UI.Pointer = t
			md.UI.EmitViewFrame(nil)
		}
	}
	if ev.Kind == "pointerdown" && ev.Hit.Kind == "edge" && ev.Hit.EdgeRow >= 0 && int(ev.Hit.EdgeRow) < len(md.RT.EdgeEndpointRowTable) && md.UI.PanelLayout().Sheet.Hit(ev.X, ev.Y).Part == Sheet.PartNone {
		e := md.RT.EdgeEndpointRowTable[ev.Hit.EdgeRow]
		if md.UI.Builder.SelectEdge(e.SrcNode, e.DstNode) {
			md.UI.EmitViewFrame(nil)
		}
	}
	if ev.Kind == "wheel" && md.UI.Builder.TreeWheel(md.UI.PanelLayout().Sheet, ev.X, ev.Y, ev.DeltaY) {
		md.UI.EmitViewFrame(nil)
		return
	}
	if ev.Kind == "wheel" && Chrome.TakeWheel(md.UI.PanelLayout(), &md.UI.OverlaysPill.Scroll, &md.UI.Sheet, ev.X, ev.Y, ev.DeltaX, ev.DeltaY, md.redraw) {
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
		"overlays": func(_ context.Context, attr byte, payload []byte) {
			Flags.ApplyUpdate(attr, payload, &md.UI.OV, &md.UI, md.persistOverlays)
		},
		"panels": func(_ context.Context, attr byte, payload []byte) {
			Panel.ApplyUpdate(attr, payload, &md.UI.PN, md.persistPanels, md.redraw)
		},
		"clock": func(_ context.Context, attr byte, payload []byte) {
			SliderPanel.ApplyUpdate(attr, payload, md.speedState(), speedSinks, md.persistSpeed, md.redraw)
		},
		"scene": func(ctx context.Context, attr byte, payload []byte) {
			applyUpdateScene(ctx, attr, payload, md, speedSinks)
		},
	}
}

// EDIT_UPDATE_KINDS_END
