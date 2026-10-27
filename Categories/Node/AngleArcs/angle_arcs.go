package AngleArcs

import (
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Vectors"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

type Vec3 = Vectors.Vec3

const (
	KindPhi       uint8 = 0
	KindTheta     uint8 = 1
	KindPhiZero   uint8 = 2
	KindThetaZero uint8 = 3
)

type Line struct {
	From, To Vec3
	Kind     uint8
}

type Label struct {
	At                   Vec3
	PhiTicks, ThetaTicks float32
}

type Arcs struct {
	Lines  []Line
	Labels []Label
}

const (
	segments     = 24
	pathRadius   = 1.5
	radiusStep   = 0.35
	chevronScale = 0.3
	labelPush    = 1.15
	zeroFrom     = 1.05
	zeroTo       = 1.35
)

func Build(selected bool, center Vec3, outer float64, ticksPerTurn int32, starts []polar.Polar) Arcs {
	var a Arcs
	if !selected {
		return a
	}
	if ticksPerTurn <= 0 {
		ticksPerTurn = 12
	}
	toTicks := float64(ticksPerTurn) / (2 * math.Pi)
	i := 0
	for _, vec := range starts {
		if vec.R < 1e-9 {
			continue
		}
		r := outer * (pathRadius + radiusStep*float64(i))
		theta := math.Mod(vec.Theta+2*math.Pi, 2*math.Pi)
		a.sweep(center, KindPhi, func(t float64) Vec3 {
			return Vec3(polar.Polar2cart(polar.Polar{R: r, Phi: t * vec.Phi}))
		})
		end := a.sweep(center, KindTheta, func(t float64) Vec3 {
			return Vec3(polar.Polar2cart(polar.Polar{R: r, Phi: vec.Phi, Theta: t * theta}))
		})
		a.chevron(center, end, KindTheta)
		a.Labels = append(a.Labels, Label{
			At:         center.Add(end[1].Scale(labelPush)),
			PhiTicks:   float32(vec.Phi * toTicks),
			ThetaTicks: float32(theta * toTicks),
		})
		i++
	}
	if i > 0 {
		up, x := Vec3{Y: outer}, Vec3{X: outer}
		a.Lines = append(a.Lines,
			Line{From: center.Add(up.Scale(zeroFrom)), To: center.Add(up.Scale(zeroTo)), Kind: KindPhiZero},
			Line{From: center.Add(x.Scale(zeroFrom)), To: center.Add(x.Scale(zeroTo)), Kind: KindThetaZero})
	}
	return a
}

func (a *Arcs) sweep(center Vec3, kind uint8, at func(t float64) Vec3) [2]Vec3 {
	prev := at(0)
	for i := 1; i <= segments; i++ {
		p := at(float64(i) / segments)
		a.Lines = append(a.Lines, Line{From: center.Add(prev), To: center.Add(p), Kind: kind})
		prev = p
	}
	return [2]Vec3{at(1 - 1.0/segments), at(1)}
}

func (a *Arcs) chevron(center Vec3, end [2]Vec3, kind uint8) {
	back, tip := end[0], end[1]
	along := tip.Sub(back)
	if along.Length() < 1e-9 {
		return
	}
	along = along.Normalize().Scale(tip.Length() * chevronScale)
	side := tip.Normalize().Scale(along.Length() * 0.6)
	a.Lines = append(a.Lines,
		Line{From: center.Add(tip), To: center.Add(tip.Sub(along).Add(side)), Kind: kind},
		Line{From: center.Add(tip), To: center.Add(tip.Sub(along).Sub(side)), Kind: kind})
}

var ValueNames = []string{
	"arcX0", "arcY0", "arcZ0", "arcX1", "arcY1", "arcZ1", "arcKind",
	"arcLabelX", "arcLabelY", "arcLabelZ", "arcLabelPhi", "arcLabelTheta",
}

type block interface {
	U8(name string, v uint8)
	F32(name string, v float32)
}

func xyz(w block, x, y, z string, v Vec3) {
	w.F32(x, float32(v.X))
	w.F32(y, float32(v.Y))
	w.F32(z, float32(v.Z))
}

func WriteValues(w block, a Arcs) {
	for _, l := range a.Lines {
		xyz(w, "arcX0", "arcY0", "arcZ0", l.From)
		xyz(w, "arcX1", "arcY1", "arcZ1", l.To)
		w.U8("arcKind", l.Kind)
	}
	for _, l := range a.Labels {
		xyz(w, "arcLabelX", "arcLabelY", "arcLabelZ", l.At)
		w.F32("arcLabelPhi", l.PhiTicks)
		w.F32("arcLabelTheta", l.ThetaTicks)
	}
}
