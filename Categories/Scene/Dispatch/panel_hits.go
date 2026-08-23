package Dispatch

import (
	"context"

	Speed "github.com/dtauraso/wirefold/Categories/Speed"

	"github.com/dtauraso/wirefold/Categories/Scene/Drag"

	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/Categories/Chrome/Pills/NodesDropdown"
	"github.com/dtauraso/wirefold/Categories/Overlay"
	"github.com/dtauraso/wirefold/Categories/Scene/Scenes"
)

func panelTookPointerDown(
	ctx context.Context,
	ev Drag.RawInputMsg,
	md *MoveDispatch,
	speedSinks SliderPanel.Sinks,
) bool {
	pl := md.UI.PanelLayout()

	if h := pl.Rules.Hit(ev.X, ev.Y); h.Kind != PolarRulesPanel.HitNone {
		applyRulesHit(ctx, md, h)
		return true
	}

	if Panel.HitRect(pl.Fit, ev.X, ev.Y) {
		home := ev
		home.Kind = "home"
		md.HandleRawInput(ctx, home)
		return true
	}

	if i := pl.Tabs.Hit(ev.X, ev.Y); i >= 0 {
		Scenes.SelectScene(&md.Scenes, i)
		return true
	}

	switch h := pl.Nodes.Hit(ev.X, ev.Y); h.Kind {
	case NodesDropdown.HitPill:
		md.UI.Nodes.Open = !md.UI.Nodes.Open
		md.UI.EmitViewFrame(nil)
		return true
	case NodesDropdown.HitRow:
		if md.UI.Nodes.RowOpen == nil {
			md.UI.Nodes.RowOpen = map[uint8]bool{}
		}
		md.UI.Nodes.RowOpen[h.KindID] = !md.UI.Nodes.RowOpen[h.KindID]
		md.UI.PlacingKind, md.UI.PlacingPending = h.KindID, true
		md.UI.EmitViewFrame(nil)
		return true
	}

	if i := pl.Speed.Hit(ev.X, ev.Y); i >= 0 {
		setClockSpeed(md, speedSinks, SliderPanel.Settings[i].Speed)
		return true
	}
	if h := pl.Overlays.Hit(ev.X, ev.Y); h.Kind != Pills.HitNone {
		applyOverlaysHit(md, h)
		return true
	}
	if h := pl.Angle.Hit(ev.X, ev.Y); h.Kind != AngleDropdown.HitNone {
		applyAngleHit(ctx, md, speedSinks, h)
		return true
	}
	switch pl.Tilt.Hit(ev.X, ev.Y) {
	case TiltPanel.ButtonStart:
		for _, col := range pl.Tilt.Columns {
			tiltVectorEdit(ctx, md, speedSinks, col.NodeRow, "start")
		}
		return true
	case TiltPanel.ButtonReset:
		for _, col := range pl.Tilt.Columns {
			tiltVectorEdit(ctx, md, speedSinks, col.NodeRow, "reset")
		}
		return true
	}
	return false
}

func placeNodeAt(md *MoveDispatch, ev *Drag.RawInputMsg) {
	if ev.RectWidth <= 0 || ev.RectHeight <= 0 {
		return
	}
	ndcX := ((ev.X-ev.RectLeft)/ev.RectWidth)*2 - 1
	ndcY := -((ev.Y-ev.RectTop)/ev.RectHeight)*2 + 1
	CreateNode(&md.Scenes, &md.UI, md.MR.NodeGeoms(), md.nearestNodeTo, md.UI.PlacingKind, ndcX, ndcY)
}

func applyOverlaysHit(md *MoveDispatch, h Pills.Hit) {
	switch h.Kind {
	case Pills.HitPillCaret, Pills.HitHeading:
		Panel.ToggleFlag(&md.UI.PN, h.Panel)
		md.UI.PersistPanels(md.UI.PN)
	case Pills.HitPillBody, Pills.HitFlag:
		Overlay.ToggleFlag(&md.UI.OV, &md.ChannelVectorsOn, &md.UI, h.Flag)
		md.UI.PersistOverlays(md.UI.OV)
	case Pills.HitCount:
		Overlay.SetCount(&md.UI.OV, &md.ChannelVectorsOn, &md.UI, h.Flags, h.Target)
		md.UI.PersistOverlays(md.UI.OV)
	}
	md.UI.EmitViewFrame(nil)
}

func applyAngleHit(ctx context.Context, md *MoveDispatch, speedSinks SliderPanel.Sinks, h AngleDropdown.Hit) {
	switch h.Kind {
	case AngleDropdown.HitPill:
		md.UI.Angle.Open = !md.UI.Angle.Open
	case AngleDropdown.HitGroup:
		if md.UI.Angle.GroupOpen == nil {
			md.UI.Angle.GroupOpen = map[int32]bool{}
		}
		md.UI.Angle.GroupOpen[h.NodeRow] = !md.UI.Angle.GroupOpen[h.NodeRow]
	case AngleDropdown.HitLatticeUp:
		setLatticePoints(md, md.UI.LatticePoints+AngleDropdown.LatticePointsStep)
	case AngleDropdown.HitLatticeDown:
		setLatticePoints(md, md.UI.LatticePoints-AngleDropdown.LatticePointsStep)
	case AngleDropdown.HitPhiUp:
		adjustTiltPhi(ctx, md, h.NodeRow, true)
	case AngleDropdown.HitPhiDown:
		adjustTiltPhi(ctx, md, h.NodeRow, false)
	}
	md.UI.EmitViewFrame(nil)
}

func setClockSpeed(md *MoveDispatch, speedSinks SliderPanel.Sinks, speed float64) {
	Speed.SetSpeedNum(int64(speed*Speed.SpeedNumScale), md.speedState(), speedSinks, md.persistSpeed, md.redraw)
}
