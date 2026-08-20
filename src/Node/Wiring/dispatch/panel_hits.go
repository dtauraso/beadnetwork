package dispatch

import (
	"context"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/wirefold/src/Chrome/Pills"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/NodesDropdown"
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodecrud"
	"github.com/dtauraso/wirefold/src/Node/Wiring/sceneswitch"
	"github.com/dtauraso/wirefold/src/Overlay"
	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
)

func panelTookPointerDown(
	ctx context.Context,
	ev inputcodec.RawInputMsg,
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
		md.HandleRawInput(ctx, home, nil)
		return true
	}

	if i := pl.Tabs.Hit(ev.X, ev.Y); i >= 0 {
		sceneswitch.SelectScene(&md.Scenes, i)
		return true
	}

	switch h := pl.Nodes.Hit(ev.X, ev.Y); h.Kind {
	case NodesDropdown.HitPill:
		md.UI.NodesOpen = !md.UI.NodesOpen
		md.UI.EmitViewFrame(nil)
		return true
	case NodesDropdown.HitRow:
		if md.UI.NodesRowOpen == nil {
			md.UI.NodesRowOpen = map[uint8]bool{}
		}
		md.UI.NodesRowOpen[h.KindID] = !md.UI.NodesRowOpen[h.KindID]
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

func placeNodeAt(md *MoveDispatch, ev *inputcodec.RawInputMsg) {
	if ev.RectWidth <= 0 || ev.RectHeight <= 0 {
		return
	}
	ndcX := ((ev.X-ev.RectLeft)/ev.RectWidth)*2 - 1
	ndcY := -((ev.Y-ev.RectTop)/ev.RectHeight)*2 + 1
	nodecrud.CreateNode(&md.Scenes, &md.UI, &md.MR, md.UI.PlacingKind, ndcX, ndcY)
}

func applyOverlaysHit(md *MoveDispatch, h Pills.Hit) {
	switch h.Kind {
	case Pills.HitPillCaret, Pills.HitHeading:
		if fn, ok := Panel.PanelToggles[h.Panel]; ok {
			fn(&md.UI.PN)
			md.Persist.Panels().Schedule(md.UI.PN)
		}
	case Pills.HitPillBody, Pills.HitFlag:
		toggleOverlayFlag(md, h.Flag)
		md.Persist.Overlays().Schedule(md.UI.OV)
	case Pills.HitCount:
		for _, flag := range h.Flags {
			read, ok := Overlay.OverlayFlagRead[flag]
			if !ok || read(&md.UI.OV) == h.Target {
				continue
			}
			toggleOverlayFlag(md, flag)
		}
		md.Persist.Overlays().Schedule(md.UI.OV)
	}
	md.UI.EmitViewFrame(nil)
}

func toggleOverlayFlag(md *MoveDispatch, flag string) {
	fn, ok := Overlay.OverlayToggles[flag]
	if !ok {
		return
	}
	fn(&md.UI.OV)
	if flag == "ruleChannels" {
		md.Inboxes.BroadcastChannelVectorsOn(md.UI.OV.RuleChannelsVisible)
	}
	if scope, ok := Overlay.OverlayFlagBreadcrumbScope[flag]; ok {
		md.UI.EmitBreadcrumb(B.RowEvent{
			Label: B.BreadcrumbPoleToggleGo, NodeRow: -1, PortRow: -1, TargetRow: -1,
			TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Value: int32(boolU8(Overlay.OverlayFlagValue[flag](&md.UI.OV))), Text: scope,
		})
	}
}

func applyAngleHit(ctx context.Context, md *MoveDispatch, speedSinks SliderPanel.Sinks, h AngleDropdown.Hit) {
	switch h.Kind {
	case AngleDropdown.HitPill:
		md.UI.AngleOpen = !md.UI.AngleOpen
	case AngleDropdown.HitGroup:
		if md.UI.AngleGroupOpen == nil {
			md.UI.AngleGroupOpen = map[int32]bool{}
		}
		md.UI.AngleGroupOpen[h.NodeRow] = !md.UI.AngleGroupOpen[h.NodeRow]
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
	divisor := int64(md.UI.ClockDivisor)
	SliderPanel.Broadcast(speedSinks, int64(speed*SliderPanel.NumScale), divisor)
	md.UI.Speed = speed
	md.Persist.Speed().Schedule(speed)
	md.UI.EmitViewFrame(nil)
}
