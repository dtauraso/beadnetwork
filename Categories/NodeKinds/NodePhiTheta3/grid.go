package NodePhiTheta3

import (
	"fmt"
	"path/filepath"

	NodeCat "github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Node/Edge/edgegeom"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

func (n *NodePhiTheta3) gridGeom() Links.GridGeom {
	kind := n.geom.Kind()
	steps := NodeCat.NodeTorusSteps(kind)
	return Links.GridGeom{
		NodeR:     NodeCat.NodeRadius(kind),
		LineSteps: func(dist float64) int { return edgegeom.EdgeStepCount(dist, steps, steps) },
	}
}

func (n *NodePhiTheta3) loadGrid(level string) {
	root, id := n.geom.PersistRoot(), n.peer.ID
	g := Links.Grid{ID: id, Level: level}
	if root != "" {
		var ok bool
		var err error
		g, ok, err = Links.ReadGrid(Links.GridPath(NodeCat.NodeDir(filepath.Join(root, level), id)))
		if err != nil {
			panic(fmt.Sprintf("NodePhiTheta3.loadGrid: node %s could not read its grid at level %q: %v — the node is the only writer of this file, so a grid it cannot read was written by something else or cut short", id, level, err))
		}
		if !ok {
			g, err = n.convertGrid(root, level)
			if err != nil {
				panic(fmt.Sprintf("NodePhiTheta3.loadGrid: node %s could not build its grid from the step link files at level %q: %v", id, level, err))
			}
		}
		g.Level = level
	}
	n.peer.SetGrid(g, n.gridGeom())
}

func (n *NodePhiTheta3) convertGrid(root, level string) (Links.Grid, error) {
	id := n.peer.ID
	levelDir := filepath.Join(root, level)
	dirs := map[int]string{}
	old := []string{NodeCat.NodeDir(levelDir, id)}
	for _, s := range Links.StepNumbers(levelDir) {
		dirs[s] = NodeCat.NodeDir(Links.StepDir(levelDir, s), id)
		old = append(old, dirs[s])
	}
	g, err := Links.ConvertGrid(level, id, dirs, n.peer.Links(), *n.peer.At, *n.peer.Ref, n.peer.S)
	if err != nil {
		return g, err
	}
	n.peer.SetGrid(g, n.gridGeom())
	n.peer.Snap()
	g = n.peer.Grid()
	if err := Links.WriteGrid(Links.GridPath(old[0]), g); err != nil {
		return g, err
	}
	n.breadcrumb("grid-converted", fmt.Sprintf("level=%q steps=%v", level, g.Steps()))
	return g, Links.RemoveOldLinks(old...)
}

func (n *NodePhiTheta3) writeGrid() {
	root := n.geom.PersistRoot()
	if root == "" {
		return
	}
	g := n.peer.Grid()
	if err := Links.WriteGrid(Links.GridPath(NodeCat.NodeDir(filepath.Join(root, g.Level), g.ID)), g); err != nil {
		n.breadcrumb("grid-persist", err.Error())
	}
}

func (n *NodePhiTheta3) showLevel(m Links.ShowStep) {
	if m.Level == n.peer.Level() {
		return
	}
	n.send(n.peer.Snap())
	n.writeGrid()
	n.loadGrid(m.Level)
}
