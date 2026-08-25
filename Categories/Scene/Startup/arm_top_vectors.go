package Startup

import (
	"github.com/dtauraso/beadnetwork/Categories/Scene/Dispatch"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
)

type topVectorHolder struct {
	scene    string
	node     string
	targetID string
}

var topVectorHolders = []topVectorHolder{
	{scene: "ring", node: "1", targetID: "3"},
}

func ArmTopVectors(md *Dispatch.MoveDispatch, scenePath string) {
	scene := Scenes.For(scenePath).Name
	for _, h := range topVectorHolders {
		if h.scene != scene {
			continue
		}
		nm, ok := md.MR.NodeGeoms()[h.node]
		if !ok {
			continue
		}
		nm.ArmTopVector(scenePath, h.targetID)
	}
}
