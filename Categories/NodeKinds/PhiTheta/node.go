package PhiTheta

import (
	"context"
	"fmt"

	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

type PhiTheta struct {
	geom *NodeCat.NodeGeometry

	In     Links.Inbox
	mesh   *Links.Mesh
	peer   *Links.Peer
	tickMs int
	showAt int64
	ctx    context.Context
}

func (n *PhiTheta) send(out []Links.Envelope) {
	for _, e := range out {
		n.mesh.Send(n.ctx, n.peer.ID, e)
	}
}

func (n *PhiTheta) breadcrumb(label, value string) {
	n.geom.Trace().Post([]NodeCat.RowEvent{{
		Kind: NodeCat.KindBreadcrumb, Label: label, Debug: 1,
		NodeRow: n.geom.Stream().NodeRow(),
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: value,
	}})
}

func (n *PhiTheta) Update(ctx context.Context) {
	n.ctx = ctx
	n.peer.SetGeom(n.nodeGeom())
	n.postTicks()
	n.publish()

	h := clock.HumanTick(n.tickMs)
	pending := false
	for {
		t := clock.NextHumanTickTimer(n.tickMs)
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case msg := <-n.In.In:
			t.Stop()
			pending = n.take(ctx, msg) || pending
		case <-t.C:
		}
		if now := clock.HumanTick(n.tickMs); now != h {
			h = now
			if pending && n.peer.Ready() && h >= n.showAt {
				n.apply()
				pending = false
			}
			n.mesh.PostTiming(Links.Timing{ID: n.peer.ID, H: h})
		}
	}
}

func (n *PhiTheta) take(ctx context.Context, msg any) bool {
	if t, ok := msg.(Links.SetTickMs); ok {
		n.tickMs = max(t.V, 1)
		n.geom.RunOnHumanTicks(n.tickMs)
		return false
	}
	if s, ok := msg.(Links.StepAll); ok {
		n.showAt = s.H + 2
	}
	out := n.peer.Handle(msg)
	n.send(out)
	picked := ""
	for _, l := range n.peer.Links() {
		if l.Picked {
			picked += l.From + "→" + l.To + " "
		}
	}
	n.breadcrumb("msg", fmt.Sprintf("%T %+v | at=%+v picked=[%s] sent=%d", msg, msg, *n.peer.At, picked, len(out)))
	return true
}

func (n *PhiTheta) apply() {
	c := n.peer.TakeChanges()
	if c.Place {
		n.geom.KindPosts().PostPlace(*n.peer.At)
	}
	if c.S {
		n.postTicks()
	}
	n.publishWith(c.Refused)
}

func (n *PhiTheta) publish() { n.publishWith("") }

func (n *PhiTheta) publishWith(refused string) {
	n.send(n.peer.Snap())
	n.writeNode()
	n.postVectors()
	n.mesh.Report(Links.Report{ID: n.peer.ID, Node: n.peer.Node(), Round: n.peer.Round(), Refused: refused})
}

func (a BuildArgs) Geom() *NodeCat.NodeGeometry {
	if a.Deps == nil {
		return nil
	}
	ng, _ := a.Deps.SelfDriveGeom(a.Name).(*NodeCat.NodeGeometry)
	return ng
}

var Builder = BuilderFor("PhiTheta",
	func(a BuildArgs) (any, error) {
		n := &PhiTheta{}
		n.geom = a.Geom()
		n.In, n.mesh = a.Inbox(), a.Mesh()
		seed := a.Seed()
		if _, ok := seed.Places[a.Name]; !ok {
			return nil, fmt.Errorf("PhiTheta: node %q has no float place (base/at-r.bin, at-phi.bin, at-theta.bin) — a links node is placed on the scene sphere by its own place, and has nowhere to stand without one", a.Name)
		}
		n.peer = Links.NewPeer(a.Name, seed)
		n.tickMs = max(seed.TickMs, 1)
		n.geom.HumanTicksOnly(n.tickMs)
		n.breadcrumb("built", fmt.Sprintf("node=%s s=%d at=%+v links=%d", a.Name, n.peer.S, *n.peer.At, len(n.peer.Links())))
		return n, nil
	})
