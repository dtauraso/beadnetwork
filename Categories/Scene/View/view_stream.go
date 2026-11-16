package View

import (
	"fmt"

	"github.com/dtauraso/beadnetwork/Categories/Scene/Preview"
)

type ViewOverlayFlags struct {
	SceneTori, ScenePoles, NodePoles, Handholds, LabelsGlobal, OverlaysVis uint8
	NodeBody, NodeRing                                                     uint8
	NodePoleSphere                                                         uint8
	AllPoleSpheres                                                         uint8
}

type ViewPanelFlags struct {
	Overlays                                                  uint8
	Node, NodeShape, NodeState, NodePoles                     uint8
	Scene, SceneGuides, ScenePoles, SceneVectors, SceneLabels uint8
}

type ViewSceneState struct {
	EditRefused   uint32
	SceneEditable uint8
	SceneKinds    uint32
}

type ViewFrameBuilder func(tick uint32, events []RowEvent)

func (ui *UIState) SetViewStream(buildFrame ViewFrameBuilder) {
	ui.ViewBuildFrame = buildFrame
}

func (ui *UIState) WriteOwnTrace() {
	ui.SetViewStream(func(tick uint32, events []RowEvent) {
		appendTrace(viewTracePath(ui.SceneRoot()), events)
	})
}

func (ui *UIState) EmitBreadcrumb(ev RowEvent) {
	ev.Kind = KindBreadcrumb
	ev.Debug = 1
	ui.EmitViewFrame([]RowEvent{ev})
}

func (ui *UIState) linksPreview(board bool) Preview.Preview {
	p := Preview.None()
	if !ui.Builder.On() || ui.NodeRowFor == nil {
		return p
	}
	for _, e := range ui.Builder.EdgeStates() {
		a, ok := ui.NodeRowFor(e.From)
		b, ok2 := ui.NodeRowFor(e.To)
		if !board || !ok || !ok2 {
			continue
		}
		st := Preview.EdgeUnpicked
		if e.Open {
			st = Preview.EdgeOpen
		} else if e.Picked {
			st = Preview.EdgePicked
		}
		p.Edges = append(p.Edges, Preview.EdgeMark{Src: a, Dst: b, State: st})
	}
	p.EdgePick = true
	from, to := ui.Builder.Preview()
	if a, ok := ui.NodeRowFor(from); ok {
		if b, ok := ui.NodeRowFor(to); ok {
			p.HotSrc, p.HotDst = a, b
		}
	}
	return p
}

func (ui *UIState) EmitViewFrame(events []RowEvent) {
	if ui.ViewBuildFrame == nil {
		return
	}
	ui.viewTick++

	ui.writeSceneColumns()
	ui.writePointerTargetColumns()

	if ui.ViewW <= 0 || ui.ViewH <= 0 {
		ui.LinksPreview.Write(ui.linksPreview(false))
		ui.ViewBuildFrame(ui.viewTick, events)
		return
	}

	pl := ui.PanelLayout()
	ui.LinksPreview.Write(ui.linksPreview(boardOf(pl.Sheet) != nil))
	ui.Slider.Write(pl.Speed, ui.Speed)
	ui.Sheet.Write(pl.Sheet)
	ui.Board.Write(boardOf(pl.Sheet), ui.Sheet.Edit)
	ui.Angle.Write(pl.Angle)
	ui.Nodes.Write(pl.Nodes, ui.EditRefused)
	ui.OverlaysPill.Write(pl.Overlays)
	ui.Fit.Write(pl.Fit)
	if rect := pl.Tabs.Strip; rect != ui.lastStrip {
		ui.lastStrip = rect
		appendTrace(viewTracePath(ui.SceneRoot()), []RowEvent{{
			Kind: KindBreadcrumb, Label: "strip-rect", Debug: 1,
			NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Text: fmt.Sprintf("x=%.4f w=%.4f tabs=%d selected=%d viewW=%.4f tick=%d",
				rect.X, rect.W, len(pl.Tabs.Tabs), ui.TabStrip.Selected, ui.ViewW, ui.viewTick),
		}})
	}
	ui.TabStrip.Write(pl.Tabs)

	ui.ViewBuildFrame(ui.viewTick, events)
}
