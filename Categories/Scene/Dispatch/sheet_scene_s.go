package Dispatch

import (
	"context"
	"fmt"
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
)

func setSceneS(ctx context.Context, md *MoveDispatch, v float64) {
	m := &md.UI.Builder
	if !whole(v) || v < 1 {
		m.Refused = fmt.Sprintf("s %v: not a whole number of at least 1, so it stays at %d", v, m.S)
		return
	}
	s := int(math.Round(v))
	if root := md.Scenes.TreeRoot; root != "" {
		if err := Sheet.WriteScalar(Scenes.SFilePath(root), int32(s)); err != nil {
			m.Refused = fmt.Sprintf("could not write s: %v", err)
			return
		}
	}
	m.S = s
	for _, id := range m.Nodes {
		md.sendLink(ctx, id, Links.SetS{V: s})
	}
}
