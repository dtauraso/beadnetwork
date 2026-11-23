package Store

import "github.com/dtauraso/beadnetwork/Categories/Node/Links"

type row[T any] struct {
	name  string
	isInt bool
	getI  func(t T) int64
	setI  func(t *T, v int64)
	getF  func(t T) float64
	setF  func(t *T, v float64)
}

func intRow[T any](name string, at func(t *T) *int) row[T] {
	return row[T]{name: name, isInt: true,
		getI: func(t T) int64 { return int64(*at(&t)) },
		setI: func(t *T, v int64) { *at(t) = int(v) }}
}

func boolRow[T any](name string, at func(t *T) *bool) row[T] {
	return row[T]{name: name, isInt: true,
		getI: func(t T) int64 { return b2i(*at(&t)) },
		setI: func(t *T, v int64) { *at(t) = v != 0 }}
}

func floatRow[T any](name string, at func(t *T) *float64) row[T] {
	return row[T]{name: name,
		getF: func(t T) float64 { return *at(&t) },
		setF: func(t *T, v float64) { *at(t) = v }}
}

func vecRows[T any](prefix string, at func(t *T) *Links.Vec3) []row[T] {
	return []row[T]{
		floatRow(prefix+" x", func(t *T) *float64 { return &at(t).X }),
		floatRow(prefix+" y", func(t *T) *float64 { return &at(t).Y }),
		floatRow(prefix+" z", func(t *T) *float64 { return &at(t).Z }),
	}
}

func b2i(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

var linkRows = []row[Links.Link]{
	intRow("order", func(l *Links.Link) *int { return &l.Order }),
	intRow(RowPhiTick, func(l *Links.Link) *int { return &l.Start.Tick.Phi }),
	intRow(RowThetaTick, func(l *Links.Link) *int { return &l.Start.Tick.Theta }),
	floatRow("r start", func(l *Links.Link) *float64 { return &l.Start.R }),
	intRow("s", func(l *Links.Link) *int { return &l.Start.S }),
	intRow("φ Δ", func(l *Links.Link) *int { return &l.Delta.Phi }),
	intRow("θ Δ", func(l *Links.Link) *int { return &l.Delta.Theta }),
	intRow("r Δ", func(l *Links.Link) *int { return &l.Delta.R }),
	intRow("φ way", func(l *Links.Link) *int { return &l.Way.Phi }),
	intRow("θ way", func(l *Links.Link) *int { return &l.Way.Theta }),
	intRow("φ nΔ", func(l *Links.Link) *int { return &l.Increment.Phi }),
	intRow("θ nΔ", func(l *Links.Link) *int { return &l.Increment.Theta }),
	intRow("r nΔ", func(l *Links.Link) *int { return &l.Increment.R }),
	intRow("φ 1", func(l *Links.Link) *int { return &l.Start.One.Phi }),
	intRow("θ 1", func(l *Links.Link) *int { return &l.Start.One.Theta }),
	intRow("r 1", func(l *Links.Link) *int { return &l.Start.One.R }),
	intRow("φ lock", func(l *Links.Link) *int { return &l.Lock.Phi }),
	intRow("θ lock", func(l *Links.Link) *int { return &l.Lock.Theta }),
	intRow("r lock", func(l *Links.Link) *int { return &l.Lock.R }),
	boolRow("picked", func(l *Links.Link) *bool { return &l.Picked }),
}

const (
	RowPhiTick   = "φ start tick"
	RowThetaTick = "θ start tick"
	rowPhiOld    = "φ start"
	rowThetaOld  = "θ start"
)

var baseRows = []row[Links.Base]{
	intRow(RowPhiTick, func(b *Links.Base) *int { return &b.Origin.Tick.Phi }),
	intRow(RowThetaTick, func(b *Links.Base) *int { return &b.Origin.Tick.Theta }),
	intRow("s", func(b *Links.Base) *int { return &b.Origin.S }),
	floatRow("r start", func(b *Links.Base) *float64 { return &b.Origin.R }),
	intRow("φ 1", func(b *Links.Base) *int { return &b.Origin.One.Phi }),
	intRow("θ 1", func(b *Links.Base) *int { return &b.Origin.One.Theta }),
	intRow("r 1", func(b *Links.Base) *int { return &b.Origin.One.R }),
	intRow("φ Δ total", func(b *Links.Base) *int { return &b.Total.Phi }),
	intRow("θ Δ total", func(b *Links.Base) *int { return &b.Total.Theta }),
	intRow("r Δ total", func(b *Links.Base) *int { return &b.Total.R }),
}

var sumRows = []row[Links.Angles]{
	floatRow("φ sum", func(s *Links.Angles) *float64 { return &s.Phi }),
	floatRow("θ sum", func(s *Links.Angles) *float64 { return &s.Theta }),
	floatRow("r sum", func(s *Links.Angles) *float64 { return &s.R }),
}

var nodeRows = append([]row[Links.Col]{
	boolRow("counted", func(c *Links.Col) *bool { return &c.Counted }),
	intRow("ref Δφ", func(c *Links.Col) *int { return &c.Ref.Phi }),
	intRow("ref Δθ", func(c *Links.Col) *int { return &c.Ref.Theta }),
	intRow("ref s", func(c *Links.Col) *int { return &c.Ref.S }),
	floatRow("place r", func(c *Links.Col) *float64 { return &c.At.R }),
	floatRow("place φ", func(c *Links.Col) *float64 { return &c.At.Phi }),
	floatRow("place θ", func(c *Links.Col) *float64 { return &c.At.Theta }),
}, append(vecRows("place", func(c *Links.Col) *Links.Vec3 { return &c.Pos }), vecRows("label anchor", func(c *Links.Col) *Links.Vec3 { return &c.Label })...)...)

var outRows = append(vecRows("edge start", func(o *Links.OutCell) *Links.Vec3 { return &o.Start }),
	append(vecRows("edge end", func(o *Links.OutCell) *Links.Vec3 { return &o.End }),
		intRow("line steps", func(o *Links.OutCell) *int { return &o.Steps }))...)
