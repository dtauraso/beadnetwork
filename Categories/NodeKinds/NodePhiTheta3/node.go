package NodePhiTheta3

import (
	"context"
	"fmt"

	clock "github.com/dtauraso/beadnetwork/Categories/Clock"
	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

type NodePhiTheta3 struct {
	geom *NodeCat.NodeGeometry

	Clock   clock.Clock
	SpeedCh <-chan float64

	In     Links.Inbox
	mesh   *Links.Mesh
	peer   *Links.Peer
	tickMs int

	k        int64
	recovers int
}

func (n *NodePhiTheta3) breadcrumb(label, value string) {
	n.geom.Trace().Post([]NodeCat.RowEvent{{
		Kind: NodeCat.KindBreadcrumb, Label: label, Debug: 1,
		NodeRow: n.geom.Stream().NodeRow(),
		PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1, Slot: -1,
		Text: value,
	}})
}

func (n *NodePhiTheta3) Update(ctx context.Context) {
	clk := n.Clock.Copy()
	clk.SpeedFrom(n.SpeedCh)
	n.geom.Clocks().Use(clk)
	n.geom.RunOnHumanTicks(n.tickMs)
	n.postTicks()
	n.publish()

	h := clock.HumanTick(n.tickMs)
	pending := false
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-n.In.In:
			pending = n.take(ctx, msg) || pending
		default:
		}
		n.k++
		if now := clock.HumanTick(n.tickMs); now != h {
			h = now
			if pending {
				n.apply()
				pending = false
			}
			n.mesh.PostTiming(Links.Timing{ID: n.peer.ID, H: h, K: n.k})
		}
	}
}

func (n *NodePhiTheta3) stepTrace(msg any, wasPlaced bool) {
	h := clock.HumanTick(n.tickMs)
	switch m := msg.(type) {
	case Links.StepAll:
		n.recovers = 0
		n.breadcrumb("step", fmt.Sprintf("round=%d h=%d k=%d placed=%v", m.Round, h, n.k, n.peer.Placed()))
	case Links.Put:
		if !m.Step {
			return
		}
		n.recovers++
		if !wasPlaced && n.peer.Placed() {
			n.breadcrumb("placed", fmt.Sprintf("round=%d h=%d k=%d recovers=%d final-from-source=%v", n.peer.Round(), h, n.k, n.recovers, m.Final))
		}
	}
}

func (n *NodePhiTheta3) take(ctx context.Context, msg any) bool {
	if t, ok := msg.(Links.SetTickMs); ok {
		n.tickMs = max(t.V, 1)
		n.geom.RunOnHumanTicks(n.tickMs)
		return false
	}
	wasPlaced := n.peer.Placed()
	for _, e := range n.peer.Handle(msg) {
		n.mesh.Send(ctx, n.peer.ID, e)
	}
	n.stepTrace(msg, wasPlaced)
	return true
}

func (n *NodePhiTheta3) apply() {
	c := n.peer.TakeChanges()
	n.breadcrumb("publish", fmt.Sprintf("round=%d h=%d k=%d place=%v links=%d placed=%v", n.peer.Round(), clock.HumanTick(n.tickMs), n.k, c.Place, len(c.Links), n.peer.Placed()))
	root, id := n.geom.PersistRoot(), n.peer.ID
	if root == "" {
		c = Links.Changes{Place: c.Place, S: c.S, Refused: c.Refused}
		if c.Place {
			n.geom.KindPosts().PostPlace(n.peer.At)
		}
		if c.S {
			n.postTicks()
		}
		n.publishWith(c.Refused)
		return
	}
	if c.Place {
		if err := NodeCat.WritePlace(root, id, n.peer.At); err != nil {
			n.breadcrumb("place-persist", err.Error())
		}
		n.geom.KindPosts().PostPlace(n.peer.At)
		n.breadcrumb("sat", fmt.Sprintf("at=%+v", n.peer.At))
	}
	if c.S {
		if err := NodeCat.WriteNodeS(root, id, n.peer.S); err != nil {
			n.breadcrumb("s-persist", err.Error())
		}
		n.postTicks()
	}
	if len(c.Links) > 0 || len(c.Dropped) > 0 {
		if err := Links.WriteLinks(NodeCat.NodeDir(root, id), n.peer.Links()); err != nil {
			n.breadcrumb("links-persist", err.Error())
		}
	}
	n.publishWith(c.Refused)
}

func (n *NodePhiTheta3) publish() { n.publishWith("") }

func (n *NodePhiTheta3) publishWith(refused string) {
	n.postVectors()
	n.mesh.Report(Links.Report{ID: n.peer.ID, S: n.peer.S, At: n.peer.At, Links: n.peer.Links(), Refused: refused})
}

func (a BuildArgs) Geom() *NodeCat.NodeGeometry {
	if a.Deps == nil {
		return nil
	}
	ng, _ := a.Deps.SelfDriveGeom(a.Name).(*NodeCat.NodeGeometry)
	return ng
}

var Builder = BuilderFor("NodePhiTheta3",
	func(a BuildArgs) (any, error) {
		n := &NodePhiTheta3{}
		n.Clock = a.Clock()
		n.SpeedCh = a.SpeedCh()
		n.geom = a.Geom()
		n.In, n.mesh = a.Inbox(), a.Mesh()
		seed := a.Seed()
		if _, ok := seed.Places[a.Name]; !ok {
			return nil, fmt.Errorf("NodePhiTheta3: node %q has no float place (base/at-r.bin, at-phi.bin, at-theta.bin) — a links node is placed on the scene sphere by its own place, and has nowhere to stand without one", a.Name)
		}
		n.peer = Links.NewPeer(a.Name, seed)
		n.tickMs = max(seed.TickMs, 1)
		n.breadcrumb("built", fmt.Sprintf("node=%s s=%d at=%+v links=%d", a.Name, n.peer.S, n.peer.At, len(n.peer.Links())))
		return n, nil
	})
