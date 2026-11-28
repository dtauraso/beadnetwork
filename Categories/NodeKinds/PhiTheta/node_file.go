package PhiTheta

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/Edge/edgegeom"
	"github.com/dtauraso/beadnetwork/Categories/Node/Files"
	"github.com/dtauraso/beadnetwork/Categories/Node/Geom"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links/Store"
)

func (n *PhiTheta) nodeGeom() Links.NodeGeom {
	kind := n.geom.Kind()
	steps := Geom.NodeTorusSteps(kind)
	return Links.NodeGeom{
		NodeR:     Geom.NodeRadius(kind),
		LineSteps: func(dist float64) int { return edgegeom.EdgeStepCount(dist, steps, steps) },
	}
}

func (n *PhiTheta) writeNode() {
	root := n.geom.PersistRoot()
	if root == "" {
		return
	}
	nd := n.peer.Node()
	if err := Store.WriteNode(Store.NodePath(Files.NodeDir(root, nd.ID)), nd); err != nil {
		n.breadcrumb("node-persist", err.Error())
	}
}
