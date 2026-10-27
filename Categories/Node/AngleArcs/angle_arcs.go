package AngleArcs

import (
	"math"

	"github.com/dtauraso/beadnetwork/Categories/Vectors"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
)

type Vec3 = Vectors.Vec3

const (
	KindPhi   uint8 = 0
	KindTheta uint8 = 1
)

type Line struct {
	From, To Vec3
	Kind     uint8
}

type Wedge struct {
	A, B, C Vec3
	Kind    uint8
}

type Label struct {
	At    Vec3
	Kind  uint8
	Ticks float32
}

type Arcs struct {
	Lines  []Line
	Wedges []Wedge
	Labels []Label
}

const (
	segments     = 24
	phiRadius    = 1.6
	thetaRadius  = 1.3
	radiusStep   = 0.35
	chevronScale = 0.3
	labelPush    = 1.15
)

func Build(center Vec3, outer float64, ticksPerTurn int32, starts []polar.Polar) Arcs {
	if ticksPerTurn <= 0 {
		ticksPerTurn = 12
	}
	toTicks := float64(ticksPerTurn) / (2 * math.Pi)
	var a Arcs
	maxPhi, maxTheta := 0.0, 0.0
	i := 0
	for _, vec := range starts {
		if vec.R < 1e-9 {
			continue
		}
		rPhi := outer * (phiRadius + radiusStep*float64(i))
		rTheta := outer * (thetaRadius + radiusStep*float64(i))
		maxPhi, maxTheta = max(maxPhi, rPhi), max(maxTheta, rTheta)
		theta := math.Mod(vec.Theta+2*math.Pi, 2*math.Pi)
		a.add(center, KindPhi, float32(vec.Phi*toTicks), func(t float64) Vec3 {
			return Vec3(polar.Polar2cart(polar.Polar{R: rPhi, Phi: t * vec.Phi, Theta: vec.Theta}))
		})
		a.add(center, KindTheta, float32(theta*toTicks), func(t float64) Vec3 {
			return Vec3{X: rTheta * math.Cos(t*theta), Z: rTheta * math.Sin(t*theta)}
		})
		i++
	}
	if i > 0 {
		a.Lines = append(a.Lines,
			Line{From: center, To: center.Add(Vec3{Y: maxPhi}), Kind: KindPhi},
			Line{From: center, To: center.Add(Vec3{X: maxTheta}), Kind: KindTheta})
	}
	return a
}

func (a *Arcs) add(center Vec3, kind uint8, ticks float32, at func(t float64) Vec3) {
	prev := at(0)
	for i := 1; i <= segments; i++ {
		p := at(float64(i) / segments)
		a.Lines = append(a.Lines, Line{From: center.Add(prev), To: center.Add(p), Kind: kind})
		a.Wedges = append(a.Wedges, Wedge{A: center, B: center.Add(prev), C: center.Add(p), Kind: kind})
		prev = p
	}
	end, back := at(1), at(1-1.0/segments)
	a.Labels = append(a.Labels, Label{At: center.Add(end.Scale(labelPush)), Kind: kind, Ticks: ticks})
	along := end.Sub(back)
	if along.Length() < 1e-9 {
		return
	}
	along = along.Normalize().Scale(end.Length() * chevronScale)
	side := end.Normalize().Scale(along.Length() * 0.6)
	a.Lines = append(a.Lines,
		Line{From: center.Add(end), To: center.Add(end.Sub(along).Add(side)), Kind: kind},
		Line{From: center.Add(end), To: center.Add(end.Sub(along).Sub(side)), Kind: kind})
}

var ValueNames = []string{
	"arcX0", "arcY0", "arcZ0", "arcX1", "arcY1", "arcZ1", "arcKind",
	"wedgeX0", "wedgeY0", "wedgeZ0", "wedgeX1", "wedgeY1", "wedgeZ1", "wedgeX2", "wedgeY2", "wedgeZ2", "wedgeKind",
	"arcLabelX", "arcLabelY", "arcLabelZ", "arcLabelKind", "arcLabelTicks",
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
	for _, t := range a.Wedges {
		xyz(w, "wedgeX0", "wedgeY0", "wedgeZ0", t.A)
		xyz(w, "wedgeX1", "wedgeY1", "wedgeZ1", t.B)
		xyz(w, "wedgeX2", "wedgeY2", "wedgeZ2", t.C)
		w.U8("wedgeKind", t.Kind)
	}
	for _, l := range a.Labels {
		xyz(w, "arcLabelX", "arcLabelY", "arcLabelZ", l.At)
		w.U8("arcLabelKind", l.Kind)
		w.F32("arcLabelTicks", l.Ticks)
	}
}
