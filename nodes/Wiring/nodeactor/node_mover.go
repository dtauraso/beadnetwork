package nodeactor

import (
	"context"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/positionfile"
	"github.com/dtauraso/wirefold/nodes/wire/clock"
)

func nodeMetaFilePath(root, id string) string {
	return filepath.Join(root, "nodes", id, "meta.json")
}

func nodeDirPath(root, id string) string {
	return filepath.Join(root, "nodes", id)
}

type newNodeMeta struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type newNodePosition struct {
	ScenePolarR     float64 `json:"scenePolarR"`
	ScenePolarTheta float64 `json:"scenePolarTheta"`
	ScenePolarPhi   float64 `json:"scenePolarPhi"`
}

func WriteNewNodeFiles(root, id, kind string, scenePolarR, theta, phi float64) error {
	dir := nodeDirPath(root, id)
	if err := os.MkdirAll(filepath.Join(dir, "edges"), 0o755); err != nil {
		return err
	}
	if err := jsonpersist.WriteJSONAtomic(nodeMetaFilePath(root, id), newNodeMeta{ID: id, Type: kind}); err != nil {
		return err
	}
	return jsonpersist.WriteJSONAtomic(positionfile.FilePath(root, id), newNodePosition{
		ScenePolarR: scenePolarR, ScenePolarTheta: theta, ScenePolarPhi: phi,
	})
}

func RemoveNodeDir(root, id string) error {
	return os.RemoveAll(nodeDirPath(root, id))
}

type NodeMover struct {
	geom *NodeGeometry

	speedCh chan float64
}

func NewNodeMover(geom *NodeGeometry) *NodeMover {
	return &NodeMover{geom: geom}
}

func (m *NodeMover) SetSpeedCh(ch chan float64) {
	m.speedCh = ch
}

func (m *NodeMover) Run(ctx context.Context) {
	g := m.geom
	if g.clocks.clockSrc != nil {
		g.clocks.clk = g.clocks.clockSrc.Copy()
	}

	if g.tr != nil {
		g.emitGeometry()
	}
	for {
		clock.ApplySpeedNonBlocking(g.clocks.clk, m.speedCh)

		for {
			progressed := false
			select {
			case <-ctx.Done():
				return
			case msg := <-g.msg.extIn:
				g.handle(msg)
				if msg.TestDone != nil {
					close(msg.TestDone)
				}
				progressed = true
			default:
			}
			for _, ch := range g.msg.neighborIn {
				select {
				case msg := <-ch:
					g.handle(msg)
					if msg.TestDone != nil {
						close(msg.TestDone)
					}
					progressed = true
				default:
				}
			}
			if !progressed {
				break
			}
		}

		outTick := g.clocks.clk.Tick()
		for _, pw := range g.outs.outWires {
			pw.DriveOneCycle(ctx, outTick)
		}

		g.flushPending()

		g.writeStreamFrame(nil)
		if err := g.clocks.clk.SleepCycle(ctx); err != nil {
			return
		}
	}
}
