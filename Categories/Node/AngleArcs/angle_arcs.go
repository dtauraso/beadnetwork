package AngleArcs

import "github.com/dtauraso/beadnetwork/Categories/Vectors"

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

type Arcs struct {
	Lines []Line
}

const (
	pathRadius   = 1.5
	radiusStep   = 0.35
	chevronScale = 0.15
	partTick     = 0.08
	zeroFrom     = 1.05
	zeroTo       = 1.35
)

type Path struct {
	Pts  []Vec3
	Ends []int
}

type Angle struct {
	At         Vec3
	Phi, Theta Path
}

func Build(selected bool, center Vec3, outer float64, angles []Angle) Arcs {
	var a Arcs
	if !selected {
		return a
	}
	i := 0
	for _, an := range angles {
		if len(an.Phi.Pts) < 2 || len(an.Theta.Pts) < 2 {
			continue
		}
		c := center.Add(an.At)
		r := outer * (pathRadius + radiusStep*float64(i))
		a.path(c, KindPhi, r, an.Phi)
		end := a.path(c, KindTheta, r, an.Theta)
		a.chevron(c, end, KindTheta)
		up, x := Vec3{Y: outer}, Vec3{X: outer}
		a.Lines = append(a.Lines,
			Line{From: c.Add(up.Scale(zeroFrom)), To: c.Add(up.Scale(zeroTo)), Kind: KindPhiZero},
			Line{From: c.Add(x.Scale(zeroFrom)), To: c.Add(x.Scale(zeroTo)), Kind: KindThetaZero})
		i++
	}
	return a
}

func (a *Arcs) path(center Vec3, kind uint8, r float64, p Path) [2]Vec3 {
	pts := p.Pts
	at := func(i int) Vec3 { return pts[i].Normalize().Scale(r) }
	for i := 1; i < len(pts); i++ {
		a.Lines = append(a.Lines, Line{From: center.Add(at(i - 1)), To: center.Add(at(i)), Kind: kind})
	}
	for _, i := range p.Ends {
		u := pts[i].Normalize()
		a.Lines = append(a.Lines, Line{From: center.Add(u.Scale(r * (1 - partTick))), To: center.Add(u.Scale(r * (1 + partTick))), Kind: kind})
	}
	return [2]Vec3{at(len(pts) - 2), at(len(pts) - 1)}
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
}
