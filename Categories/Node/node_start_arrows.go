package Node

import (
	"github.com/dtauraso/beadnetwork/Categories/Node/TiltVectors"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

func TipIndex(center, sceneCenter Vec3, frame Frame, vec polarindex.Offset, sc polarindex.SceneConstants) polarindex.Index {
	tip := center.Add(frame.Apply(Vec3(polar.Polar2cart(polarindex.OffsetToPolar(vec, sc)))))
	return polarindex.MeasureIndex(polar.Cart2polar(polar.Vec3(tip.Sub(sceneCenter))), sc)
}

func StartFrameOf(frame Frame, vec polarindex.Offset, sc polarindex.SceneConstants) Frame {
	p := polarindex.OffsetToPolar(vec, sc)
	return frame.orIdentity().Then(DirFrame(p.Phi, p.Theta))
}

func startArrows(g NodeGeom, center Vec3, frame Frame, starts []polarindex.Offset) []TiltVectors.TiltArrow {
	sc := g.SceneConstants
	size := NodeRadius(g.Kind)
	var out []TiltVectors.TiltArrow
	for _, off := range starts {
		tip := WorldPosAt(g.SceneCenter, TipIndex(center, g.SceneCenter, frame, off, sc), sc)
		if a, ok := TiltVectors.ArrowBetween(TiltVectors.Vec3(center), TiltVectors.Vec3(tip), size, TiltVectors.ArrowStart); ok {
			out = append(out, a)
		}
	}
	return out
}
