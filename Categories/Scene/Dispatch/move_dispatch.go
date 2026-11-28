package Dispatch

import (
	"context"
	"sync"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder"
	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	"github.com/dtauraso/beadnetwork/Categories/Node"
	edge "github.com/dtauraso/beadnetwork/Categories/Node/Edge"
	"github.com/dtauraso/beadnetwork/Categories/Node/Geom"
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

func NewMoveDispatch(geoms map[string]Geom.NodeGeom, edgeEndpoints map[string]edge.EdgeEndpoints, nodeOrder, edgeOrder []string, clk clock.Clock, speedSinks *SliderPanel.Sinks, rowCount int, constants polarindex.SceneConstants) (*MoveDispatch, error) {
	nodeOrder, edgeOrder = resolveSeedOrders(geoms, edgeEndpoints, nodeOrder, edgeOrder)

	md := &MoveDispatch{}
	md.MR = NewMovers()
	initMoveDispatchUIDefaults(md)

	if err := md.buildGeomSeeds(geoms, edgeEndpoints, nodeOrder, edgeOrder); err != nil {
		return nil, err
	}
	md.buildNodeMovers(geoms, clk, constants)
	md.wireMutualPairs(edgeEndpoints)
	md.buildEdgeTable(edgeEndpoints)
	md.wireNodeEdgeIDs()
	md.buildRowTables(rowCount)
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
	if dir := md.UI.Builder.Root; dir != "" {
		if err := md.UI.Builder.WriteSteps(); err != nil {
			View.LogPersistErr("sheet_step_persist", dir, err)
		}
	}
}

func (md *MoveDispatch) ShowStep(ctx context.Context) {
	m := &md.UI.Builder
	if err := m.SaveOpen(); err != nil {
		View.LogPersistErr("sync_open", m.Scene, err)
	}
	m.ReadTree()
	md.send(ctx, m.Sync())
}

func (md *MoveDispatch) send(ctx context.Context, es []Links.Envelope) {
	if len(es) == 0 {
		return
	}
	for _, e := range es {
		md.sendLink(ctx, e.To, e.Msg)
	}
	md.Settle(ctx)
}

func (md *MoveDispatch) runStepHead(ctx context.Context, a Sheet.Act, sign int, typed *float64) {
	m := &md.UI.Builder
	if typed != nil && a.Op == Builder.OpAllScale {
		m.StepHead(a.I).Scale = max(1, int(*typed))
		md.persistStep()
		return
	}
	if md.runPlay(ctx, a) {
		return
	}
	refused, starts, err := m.StepHeadEdit(a, sign)
	if err != nil {
		View.LogPersistErr("sheet_step_amount", m.Root, err)
	}
	md.persistStep()
	if refused != "" {
		m.Refused = refused
	}
	md.ShowStep(ctx)
	md.send(ctx, starts)
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
	before := m.TimeKey()
	for _, t := range md.Links.TakeTimings() {
		m.TakeTiming(t)
	}
	m.Now = clock.HumanTick(m.TickMs)
	played := md.playTick(context.TODO())
	if len(rs) > 0 || m.TimeKey() != before || played {
		md.UI.EmitViewFrame(nil)
	}
}

func (md *MoveDispatch) runFolderPlay(ctx context.Context) {
	m := &md.UI.Builder
	if !md.stopPlay(ctx) {
		md.send(ctx, m.StartFolder(clock.HumanTick(m.TickMs)))
	}
}

func (md *MoveDispatch) runPlay(ctx context.Context, a Sheet.Act) bool {
	if a.Op != Builder.OpAllPlay {
		return false
	}
	m := &md.UI.Builder
	if !md.stopPlay(ctx) {
		md.send(ctx, m.StartPlay(a.I, clock.HumanTick(m.TickMs)))
	}
	return true
}

func (md *MoveDispatch) playTick(ctx context.Context) bool {
	m := &md.UI.Builder
	if es, ok := m.NextFolderStep(m.Now); ok {
		md.send(ctx, es)
		return true
	}
	edits, ok := m.PlayTick(m.Now)
	if !ok {
		return false
	}
	md.send(ctx, edits)
	return true
}

func (md *MoveDispatch) stopPlay(ctx context.Context) bool {
	end, was := md.UI.Builder.StopPlay()
	md.send(ctx, end)
	return was
}
