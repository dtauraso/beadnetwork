package stdinreader

import (
	"context"
	"fmt"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	"math"
	"strconv"

	"github.com/dtauraso/wirefold/src/Node/rowevent"

	"github.com/dtauraso/wirefold/src/Chrome/Pills/NodesDropdown"
	"github.com/dtauraso/wirefold/src/Node/Wiring/angledropdown"
	"github.com/dtauraso/wirefold/src/Node/Wiring/dispatch"
	"github.com/dtauraso/wirefold/src/Node/Wiring/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/Wiring/movemsg"
	"github.com/dtauraso/wirefold/src/Node/Wiring/rulenode"
	"github.com/dtauraso/wirefold/src/Node/Wiring/scenepersist"
	"github.com/dtauraso/wirefold/src/Node/Wiring/sceneswitch"
	B "github.com/dtauraso/wirefold/src/schema/buffer-layout"
)

func applyUpdateClock(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, speedSinks SliderPanel.Sinks) {
	if h, ok := clockAttrHandlers[msg.Attr]; ok {
		h(msg, md, speedSinks)
	}
}

func tiltVectorEdit(ctx context.Context, md *dispatch.MoveDispatch, speedSinks SliderPanel.Sinks, row int32, attr string) {
	id := strconv.Itoa(int(row) + 1)
	if _, ok := md.MR.NodeGeoms()[id]; !ok {
		return
	}
	SliderPanel.Broadcast(speedSinks, scenepersist.SliderNum(md.UI.Speed), int64(md.UI.ClockDivisor))
	if attr == "start" {
		md.Inboxes.SendTiltEdit(ctx, id, movemsg.TiltEditMsg{Start: true})
		return
	}
	if md.Inboxes.SendTiltEdit(ctx, id, movemsg.TiltEditMsg{Reset: true}) {
		return
	}
	md.MR.SendMove(ctx, id, movemsg.Msg{Kind: movemsg.KindTiltVectorReset, NodeID: id})
}

func applyUpdateTiltVector(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, speedSinks SliderPanel.Sinks) {
	if md == nil || (msg.Attr != "phi" && msg.Attr != "reset" && msg.Attr != "start") {
		return
	}
	id := strconv.Itoa(msg.Num + 1)
	if _, ok := md.MR.NodeGeoms()[id]; !ok {
		return
	}
	if msg.Attr == "reset" || msg.Attr == "start" {
		tiltVectorEdit(ctx, md, speedSinks, int32(msg.Num), msg.Attr)
		return
	}

	adjustTiltPhi(ctx, md, int32(msg.Num), msg.Flag == "up")
}

func adjustTiltPhi(ctx context.Context, md *dispatch.MoveDispatch, row int32, up bool) {
	id := strconv.Itoa(int(row) + 1)
	if _, ok := md.MR.NodeGeoms()[id]; !ok {
		return
	}
	if md.Inboxes.SendTiltEdit(ctx, id, movemsg.TiltEditMsg{Up: up}) {
		return
	}
	md.MR.SendMove(ctx, id, movemsg.Msg{Kind: movemsg.KindTiltVectorAngle, NodeID: id, Bool: up})
}

func setLatticePoints(md *dispatch.MoveDispatch, points int32) {
	if points < angledropdown.LatticePointsMin || points > angledropdown.LatticePointsMax || points%4 != 0 {
		return
	}
	md.UI.LatticePoints = points
	md.Persist.Lattice().Schedule(points)
	md.Inboxes.BroadcastLatticePoints(points)
}

func applyUpdateScene(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, speedSinks SliderPanel.Sinks) {
	if md == nil {
		return
	}
	switch msg.Attr {
	case "selected":
		sceneswitch.SelectScene(&md.Scenes, int(msg.Num))
	case "latticePoints":
		setLatticePoints(md, int32(msg.Num))
	case "viewport":
		md.UI.ViewW = msg.X
		md.UI.ViewH = msg.Y
		md.UI.EmitBreadcrumb(rowevent.RowEvent{
			Label: B.BreadcrumbViewport, NodeRow: -1, PortRow: -1, TargetRow: -1,
			TargetPortRow: -1, EdgeRow: -1, Slot: -1,
			Value: int32(md.UI.FovDeg()),
			Text:  fmt.Sprintf("%.0fx%.0f", msg.X, msg.Y),
		})
	case "create":

		NodesDropdown.CreateNode(&md.Scenes, &md.UI, &md.MR, uint8(msg.Num), msg.X, msg.Y)
	case "delete":

		NodesDropdown.DeleteNode(&md.Scenes, &md.UI, &md.RT, msg.Num)
	}
}

func applyUpdateOverlays(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, speedSinks SliderPanel.Sinks) {
	if md == nil {
		return
	}
	if h, ok := overlayAttrHandlers[msg.Attr]; ok {
		h(msg, md)
	}

	md.Persist.Overlays().Schedule(md.UI.OV)
}

var nodeAttrHandlers = map[string]func(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch){
	"dragPhi": func(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch) {
		sendRuleEdit(ctx, md, msg.Num, rulenode.Edit{Kind: rulenode.EditPhiToggle})
	},
	"dragMaxTheta": func(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch) {
		var maxTheta *float64
		if msg.X >= 0 {
			radians := msg.X * math.Pi
			maxTheta = &radians
		}
		sendRuleEdit(ctx, md, msg.Num, rulenode.Edit{Kind: rulenode.EditMaxTheta, MaxTheta: maxTheta})
	},
	"dragActive": func(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch) {
		sendRuleEdit(ctx, md, msg.Num, rulenode.Edit{Kind: rulenode.EditActiveToggle})
	},
	"dragR": func(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch) {
		sendRuleEdit(ctx, md, msg.Num, rulenode.Edit{Kind: rulenode.EditRToggle})
	},
	"selfDragR": func(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch) {
		sendRuleEdit(ctx, md, msg.Num, rulenode.Edit{Kind: rulenode.EditSelfRToggle})
	},
	"selfDragPhi": func(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch) {
		sendRuleEdit(ctx, md, msg.Num, rulenode.Edit{Kind: rulenode.EditSelfPhiToggle})
	},
	"selfDragMaxTheta": func(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch) {
		var maxTheta *float64
		if msg.X >= 0 {
			radians := msg.X * math.Pi
			maxTheta = &radians
		}
		sendRuleEdit(ctx, md, msg.Num, rulenode.Edit{Kind: rulenode.EditSelfMaxTheta, MaxTheta: maxTheta})
	},
	"selfDragActive": func(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch) {
		sendRuleEdit(ctx, md, msg.Num, rulenode.Edit{Kind: rulenode.EditSelfActiveToggle})
	},
	"kindActive": func(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch) {
		row := msg.Num
		if row < 0 || row >= len(md.Rules.KindTogglesByNodeRow) {
			panic(fmt.Sprintf(
				"kindActive: node row %d is outside the %d rows the tree declares, so a kind-rule toggle names an "+
					"entity the row space has no slot for — the webview and the loaded tree disagree about how many "+
					"nodes exist", row, len(md.Rules.KindTogglesByNodeRow)))
		}
		toggle := md.Rules.KindTogglesByNodeRow[row]
		if toggle == nil {
			return
		}
		select {
		case toggle <- struct{}{}:
		case <-ctx.Done():
		}
	},
}

func sendRuleEdit(ctx context.Context, md *dispatch.MoveDispatch, row int, edit rulenode.Edit) {
	if row < 0 || row >= len(md.Rules.EditsByNodeRow) {
		panic(fmt.Sprintf(
			"sendRuleEdit: node row %d is outside the %d rows the tree declares, so a rule edit names an entity "+
				"the row space has no slot for — the webview and the loaded tree disagree about how many nodes exist",
			row, len(md.Rules.EditsByNodeRow)))
	}
	edits := md.Rules.EditsByNodeRow[row]
	if edits == nil {
		return
	}
	select {
	case edits <- edit:
	case <-ctx.Done():
	}
}

func applyUpdateNode(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, speedSinks SliderPanel.Sinks) {
	if md == nil {
		return
	}
	if h, ok := nodeAttrHandlers[msg.Attr]; ok {
		h(ctx, msg, md)
	}
}

func applyUpdatePanels(ctx context.Context, msg inputcodec.StdinMsg, md *dispatch.MoveDispatch, speedSinks SliderPanel.Sinks) {
	if md == nil {
		return
	}
	if h, ok := panelAttrHandlers[msg.Attr]; ok {
		h(msg, md)
	}

	md.Persist.Panels().Schedule(md.UI.PN)
	md.UI.EmitViewFrame(nil)
}
