package PhiTheta

import (
	"fmt"
	"path/filepath"

	"github.com/dtauraso/beadnetwork/Categories/Node/Edge/edgegeom"
	"github.com/dtauraso/beadnetwork/Categories/Node/Files"
	"github.com/dtauraso/beadnetwork/Categories/Node/Geom"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links/Store"
)

func (n *PhiTheta) gridGeom() Links.GridGeom {
	kind := n.geom.Kind()
	steps := Geom.NodeTorusSteps(kind)
	return Links.GridGeom{
		NodeR:     Geom.NodeRadius(kind),
		LineSteps: func(dist float64) int { return edgegeom.EdgeStepCount(dist, steps, steps) },
	}
}

func (n *PhiTheta) loadGrid(level string) {
	root, id := n.geom.PersistRoot(), n.peer.ID
	g := Links.Grid{ID: id, Level: level}
	if root != "" {
		var ok bool
		var err error
		g, ok, err = Store.ReadGrid(Store.GridPath(Files.NodeDir(filepath.Join(root, level), id)))
		if err != nil {
			panic(fmt.Sprintf("PhiTheta.loadGrid: node %s could not read its grid at level %q: %v — the node is the only writer of this file, so a grid it cannot read was written by something else or cut short", id, level, err))
		}
		if !ok {
			g, err = n.convertGrid(root, level)
			if err != nil {
				panic(fmt.Sprintf("PhiTheta.loadGrid: node %s could not build its grid from the step link files at level %q: %v", id, level, err))
			}
		}
		g.Level = level
	}
	n.peer.SetGrid(g, n.gridGeom())
}

func (n *PhiTheta) convertGrid(root, level string) (Links.Grid, error) {
	id := n.peer.ID
	levelDir := filepath.Join(root, level)
	dirs := map[int]string{}
	old := []string{Files.NodeDir(levelDir, id)}
	for _, s := range Store.StepNumbers(levelDir) {
		dirs[s] = Files.NodeDir(Store.StepDir(levelDir, s), id)
		old = append(old, dirs[s])
	}
	g, err := Store.ConvertGrid(level, id, dirs, n.peer.Links(), *n.peer.At, *n.peer.Ref, n.peer.S)
	if err != nil {
		return g, err
	}
	n.peer.SetGrid(g, n.gridGeom())
	n.peer.Snap()
	g = n.peer.Grid()
	if err := Store.WriteGrid(Store.GridPath(old[0]), g); err != nil {
		return g, err
	}
	n.breadcrumb("grid-converted", fmt.Sprintf("level=%q steps=%v", level, g.Steps()))
	return g, Store.RemoveOldLinks(old...)
}

func (n *PhiTheta) writeGrid() {
	root := n.geom.PersistRoot()
	if root == "" {
		return
	}
	g := n.peer.Grid()
	if err := Store.WriteGrid(Store.GridPath(Files.NodeDir(filepath.Join(root, g.Level), g.ID)), g); err != nil {
		n.breadcrumb("grid-persist", err.Error())
	}
}

func (n *PhiTheta) showLevel(m Links.ShowStep) {
	if m.Level == n.peer.Level() {
		return
	}
	n.send(n.peer.Snap())
	n.writeGrid()
	n.loadGrid(m.Level)
}
