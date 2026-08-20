package stdinreader

import (
	"context"

	"github.com/dtauraso/wirefold/src/Node/rowevent"

	"github.com/dtauraso/wirefold/src/Node/Wiring/angledropdown"
	"github.com/dtauraso/wirefold/src/Node/Wiring/dispatch"
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/nodesdropdown"
	"github.com/dtauraso/wirefold/src/Node/Wiring/overlayspanel"
	"github.com/dtauraso/wirefold/src/Node/Wiring/panelstack"
	"github.com/dtauraso/wirefold/src/Node/Wiring/rulespanel"
	"github.com/dtauraso/wirefold/src/Node/Wiring/sceneswitch"
	"github.com/dtauraso/wirefold/src/Node/Wiring/speedpanel"
	"github.com/dtauraso/wirefold/src/Node/Wiring/tiltpanel"
	"github.com/dtauraso/wirefold/src/Chrome/NodesDropdown"
	"github.com/dtauraso/wirefold/src/Chrome/OverlaysDropdown"
	"github.com/dtauraso/wirefold/src/Chrome/SliderPanel"
	T "github.com/dtauraso/wirefold/src/Trace"
)

func panelTookPointerDown(
	ctx context.Context,
	ev inputcodec.RawInputMsg,
	md *dispatch.MoveDispatch,
	tr *T.Trace,
	speedSinks SliderPanel.Sinks,
) bool {
	pl := md.UI.PanelLayout()

	if h := pl.Rules.Hit(ev.X, ev.Y); h.Kind != rulespanel.HitNone {
		applyRulesHit(ctx, md, h)
		return true
	}

	if panelstack.HitRect(pl.Fit, ev.X, ev.Y) {
		home := ev
		home.Kind = "home"
		md.HandleRawInput(ctx, home, nil, tr)
		return true
	}

	if i := pl.Tabs.Hit(ev.X, ev.Y); i >= 0 {
		sceneswitch.SelectScene(&md.Scenes, i)
		return true
	}

	switch h := pl.Nodes.Hit(ev.X, ev.Y); h.Kind {
	case nodesdropdown.HitPill:
		md.UI.NodesOpen = !md.UI.NodesOpen
		md.UI.EmitViewFrame(nil)
		return true
	case nodesdropdown.HitRow:
		if md.UI.NodesRowOpen == nil {
			md.UI.NodesRowOpen = map[uint8]bool{}
		}
		md.UI.NodesRowOpen[h.KindID] = !md.UI.NodesRowOpen[h.KindID]
		md.UI.PlacingKind, md.UI.PlacingPending = h.KindID, true
		md.UI.EmitViewFrame(nil)
		return true
	}

	if i := pl.Speed.Hit(ev.X, ev.Y); i >= 0 {
		setClockSpeed(md, speedSinks, speedpanel.Settings[i].Speed)
		return true
	}
	if h := pl.Overlays.Hit(ev.X, ev.Y); h.Kind != overlayspanel.HitNone {
		applyOverlaysHit(md, tr, h)
		return true
	}
	if h := pl.Angle.Hit(ev.X, ev.Y); h.Kind != angledropdown.HitNone {
		applyAngleHit(ctx, md, speedSinks, h)
		return true
	}
	switch pl.Tilt.Hit(ev.X, ev.Y) {
	case tiltpanel.ButtonStart:
		for _, col := range pl.Tilt.Columns {
			tiltVectorEdit(ctx, md, speedSinks, col.NodeRow, "start")
		}
		return true
	case tiltpanel.ButtonReset:
		for _, col := range pl.Tilt.Columns {
			tiltVectorEdit(ctx, md, speedSinks, col.NodeRow, "reset")
		}
		return true
	}
	return false
}

func placeNodeAt(md *dispatch.MoveDispatch, ev *inputcodec.RawInputMsg, tr *T.Trace) {
	if ev.RectWidth <= 0 || ev.RectHeight <= 0 {
		return
	}
	ndcX := ((ev.X-ev.RectLeft)/ev.RectWidth)*2 - 1
	ndcY := -((ev.Y-ev.RectTop)/ev.RectHeight)*2 + 1
	NodesDropdown.CreateNode(&md.Scenes, &md.UI, &md.MR, md.UI.PlacingKind, ndcX, ndcY, tr)
}

func applyOverlaysHit(md *dispatch.MoveDispatch, tr *T.Trace, h overlayspanel.Hit) {
	switch h.Kind {
	case overlayspanel.HitPillCaret, overlayspanel.HitHeading:
		if fn, ok := OverlaysDropdown.PanelToggles[h.Panel]; ok {
			fn(&md.UI.PN)
			md.Persist.Panels().Schedule(md.UI.PN)
		}
	case overlayspanel.HitPillBody, overlayspanel.HitFlag:
		toggleOverlayFlag(md, tr, h.Flag)
		md.Persist.Overlays().Schedule(md.UI.OV)
	case overlayspanel.HitCount:
		for _, flag := range h.Flags {
			read, ok := OverlaysDropdown.OverlayFlagRead[flag]
			if !ok || read(&md.UI.OV) == h.Target {
				continue
			}
			toggleOverlayFlag(md, tr, flag)
		}
		md.Persist.Overlays().Schedule(md.UI.OV)
	}
	md.UI.EmitViewFrame(nil)
}

func toggleOverlayFlag(md *dispatch.MoveDispatch, tr *T.Trace, flag string) {
	fn, ok := OverlaysDropdown.OverlayToggles[flag]
	if !ok {
		return
	}
	fn(&md.UI.OV, tr)
	if flag == "ruleChannels" {
		md.Inboxes.BroadcastChannelVectorsOn(md.UI.OV.RuleChannelsVisible)
	}
	if scope, ok := OverlaysDropdown.OverlayFlagBreadcrumbScope[flag]; ok {
		md.UI.EmitBreadcrumb(rowevent.RowEvent{
			Label: T.BreadcrumbPoleToggleGo, NodeRow: -1, PortRow: -1, TargetRow: -1,
			TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Value: int32(boolU8(OverlaysDropdown.OverlayFlagValue[flag](&md.UI.OV))), Text: scope,
		})
	}
}

func applyAngleHit(ctx context.Context, md *dispatch.MoveDispatch, speedSinks SliderPanel.Sinks, h angledropdown.Hit) {
	switch h.Kind {
	case angledropdown.HitPill:
		md.UI.AngleOpen = !md.UI.AngleOpen
	case angledropdown.HitGroup:
		if md.UI.AngleGroupOpen == nil {
			md.UI.AngleGroupOpen = map[int32]bool{}
		}
		md.UI.AngleGroupOpen[h.NodeRow] = !md.UI.AngleGroupOpen[h.NodeRow]
	case angledropdown.HitLatticeUp:
		setLatticePoints(md, md.UI.LatticePoints+angledropdown.LatticePointsStep)
	case angledropdown.HitLatticeDown:
		setLatticePoints(md, md.UI.LatticePoints-angledropdown.LatticePointsStep)
	case angledropdown.HitPhiUp:
		adjustTiltPhi(ctx, md, h.NodeRow, true)
	case angledropdown.HitPhiDown:
		adjustTiltPhi(ctx, md, h.NodeRow, false)
	}
	md.UI.EmitViewFrame(nil)
}

func setClockSpeed(md *dispatch.MoveDispatch, speedSinks SliderPanel.Sinks, speed float64) {
	divisor := int64(md.UI.ClockDivisor)
	SliderPanel.Broadcast(speedSinks, int64(speed*SliderPanel.NumScale), divisor)
	md.UI.Speed = speed
	md.Persist.Speed().Schedule(speed)
	md.UI.EmitViewFrame(nil)
}
