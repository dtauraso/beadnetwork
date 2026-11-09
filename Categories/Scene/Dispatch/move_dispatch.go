package Dispatch

import (
	"context"
	"fmt"
	"sync"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	"github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Node/ChannelVectors"
	edge "github.com/dtauraso/beadnetwork/Categories/Node/Edge"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
	SceneB "github.com/dtauraso/beadnetwork/Categories/Scene"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
	"github.com/dtauraso/beadnetwork/Categories/Scene/View"
	Flags "github.com/dtauraso/beadnetwork/Categories/Scene/View/Flags"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type MoveDispatch struct {
	MR Movers

	GS SceneB.GeomSeeds

	Sw StreamWiring

	UI View.UIState

	Mover Node.NodeMover

	Scenes Scenes.SceneSwitch

	RT RowTables

	Inboxes TiltVectors.TiltEditInboxes

	Links *Links.Mesh

	ChannelVectorsOn ChannelVectors.OnSwitch

	Rules Node.RuleChannels
}

func (md *MoveDispatch) Start(ctx context.Context) *sync.WaitGroup {
	return md.MR.Start(ctx)
}

func initMoveDispatchUIDefaults(md *MoveDispatch) {
	md.UI.OV = Flags.DefaultOverlayState()
	md.UI.PN = Panel.DefaultPanelState()
	md.UI.Speed = 1
	md.UI.ClockDivisor = 1
	md.UI.LatticePoints = AngleDropdown.DefaultLatticePoints
}

func (md *MoveDispatch) bindUIClosures() {
	md.UI.NodeRowFor = md.RT.NodeRowFor
}

func NewMoveDispatch(geoms map[string]Node.NodeGeom, edgeEndpoints map[string]edge.EdgeEndpoints, nodeOrder, edgeOrder []string, clk clock.Clock, speedSinks *SliderPanel.Sinks, rowCount int, constants polarindex.SceneConstants, polarRules bool) (*MoveDispatch, error) {
	nodeOrder, edgeOrder = resolveSeedOrders(geoms, edgeEndpoints, nodeOrder, edgeOrder)

	md := &MoveDispatch{}
	md.MR = NewMovers()
	initMoveDispatchUIDefaults(md)

	if err := md.buildGeomSeeds(geoms, edgeEndpoints, nodeOrder, edgeOrder); err != nil {
		return nil, err
	}
	md.buildNodeMovers(geoms, clk, constants)
	if polarRules {
		md.wireRuleMesh()
	}
	md.wireMutualPairs(edgeEndpoints)
	md.buildEdgeTable(edgeEndpoints)
	md.wireNodeEdgeIDs()
	md.buildRowTables(rowCount)
	md.wireRuleEditRows()
	md.bindUIClosures()

	return md, nil
}

func (md *MoveDispatch) sendLink(ctx context.Context, to string, msg any) {
	if md.Links == nil {
		return
	}
	md.Links.Send(ctx, "", Links.Envelope{To: to, Msg: msg})
}

func (md *MoveDispatch) persistStep() {
	if dir := md.UI.Builder.StepDir; dir != "" {
		if err := md.UI.Builder.WriteSteps(); err != nil {
			View.LogPersistErr("sheet_step_persist", dir, err)
		}
	}
}

func (md *MoveDispatch) syncStepPicks(ctx context.Context, step int) {
	unpick, pick := md.UI.Builder.StepPickChanges(step)
	for _, l := range unpick {
		md.sendLink(ctx, l.To, Links.Unpick{From: l.From})
	}
	for _, l := range pick {
		md.sendLink(ctx, l.To, Links.PickReq{From: l.From})
	}
}

func (md *MoveDispatch) runStepHead(ctx context.Context, a Sheet.Act, sign int) {
	m := &md.UI.Builder
	was := m.StepDeltas(a.I)
	refused, moved := m.StepHeadEdit(a, sign)
	m.Refused = refused
	md.persistStep()
	if len(moved) > 0 {
		md.syncStepPicks(ctx, a.I)
	}
	for _, e := range moved {
		md.sendStepLink(ctx, e, was, a.Op != Builder.OpAllOne)
	}
}

func (md *MoveDispatch) sendStepLink(ctx context.Context, e Builder.StepEdge, was map[Builder.StepEdge]Links.Ints, fromDrawn bool) {
	m := &md.UI.Builder
	l := m.StepLinks[e]
	if l == nil {
		return
	}
	start := l.Start
	if fromDrawn {
		if v, ok := m.TurnFromDrawn(e, was[e]); ok {
			start = v
		}
	}
	md.sendLink(ctx, e.To, Links.SetLink{From: e.From, Start: start, Delta: l.Delta, Increment: l.Increment})
}

func frameSIsLargestNodeS(md *MoveDispatch) {
	m := &md.UI.Builder
	top := 0
	for _, id := range m.Nodes {
		top = max(top, m.Reports[id].S)
	}
	if top < 1 || top == m.S {
		return
	}
	m.S = top
	if root := md.Scenes.TreeRoot; root != "" {
		if err := Sheet.WriteScalar(Scenes.SFilePath(root), int32(top)); err != nil {
			m.Refused = fmt.Sprintf("could not write s: %v", err)
		}
	}
}

func DrainLinkReports(md *MoveDispatch) {
	if md == nil || md.Links == nil {
		return
	}
	m := &md.UI.Builder
	rs := md.Links.TakeReports()
	for _, r := range rs {
		m.Take(r)
	}
	if len(rs) > 0 {
		frameSIsLargestNodeS(md)
	}
	if len(rs) > 0 && m.StepDir != "" && m.SeedStepLinks() {
		md.persistStep()
	}
	before := m.TimeKey()
	for _, t := range md.Links.TakeTimings() {
		m.TakeTiming(t)
	}
	m.Now = clock.HumanTick(m.TickMs)
	if len(rs) > 0 || m.TimeKey() != before {
		md.UI.EmitViewFrame(nil)
	}
}
