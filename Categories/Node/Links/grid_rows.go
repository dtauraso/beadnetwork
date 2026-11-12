package Links

type row[T any] struct {
	name string
	get  func(t T) float64
	set  func(t *T, v float64)
}

func intRow[T any](name string, at func(t *T) *int) row[T] {
	return row[T]{name, func(t T) float64 { return float64(*at(&t)) }, func(t *T, v float64) { *at(t) = int(v) }}
}

func floatRow[T any](name string, at func(t *T) *float64) row[T] {
	return row[T]{name, func(t T) float64 { return *at(&t) }, func(t *T, v float64) { *at(t) = v }}
}

func vecRows[T any](prefix string, at func(t *T) *Vec3) []row[T] {
	return []row[T]{
		floatRow(prefix+" x", func(t *T) *float64 { return &at(t).X }),
		floatRow(prefix+" y", func(t *T) *float64 { return &at(t).Y }),
		floatRow(prefix+" z", func(t *T) *float64 { return &at(t).Z }),
	}
}

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

var linkRows = []row[Link]{
	intRow("order", func(l *Link) *int { return &l.Order }),
	floatRow("φ start", func(l *Link) *float64 { return &l.Start.Phi }),
	floatRow("θ start", func(l *Link) *float64 { return &l.Start.Theta }),
	floatRow("r start", func(l *Link) *float64 { return &l.Start.R }),
	intRow("s", func(l *Link) *int { return &l.Start.S }),
	intRow("φ Δ", func(l *Link) *int { return &l.Delta.Phi }),
	intRow("θ Δ", func(l *Link) *int { return &l.Delta.Theta }),
	intRow("r Δ", func(l *Link) *int { return &l.Delta.R }),
	intRow("φ nΔ", func(l *Link) *int { return &l.Increment.Phi }),
	intRow("θ nΔ", func(l *Link) *int { return &l.Increment.Theta }),
	intRow("r nΔ", func(l *Link) *int { return &l.Increment.R }),
	intRow("φ 1", func(l *Link) *int { return &l.Start.One.Phi }),
	intRow("θ 1", func(l *Link) *int { return &l.Start.One.Theta }),
	intRow("r 1", func(l *Link) *int { return &l.Start.One.R }),
	intRow("φ lock", func(l *Link) *int { return &l.Lock.Phi }),
	intRow("θ lock", func(l *Link) *int { return &l.Lock.Theta }),
	intRow("r lock", func(l *Link) *int { return &l.Lock.R }),
	{"picked", func(l Link) float64 { return b2f(l.Picked) }, func(l *Link, v float64) { l.Picked = v != 0 }},
}

var baseRows = []row[Base]{
	floatRow("φ start", func(b *Base) *float64 { return &b.Origin.Phi }),
	floatRow("θ start", func(b *Base) *float64 { return &b.Origin.Theta }),
	floatRow("r start", func(b *Base) *float64 { return &b.Origin.R }),
	intRow("φ 1", func(b *Base) *int { return &b.Origin.One.Phi }),
	intRow("θ 1", func(b *Base) *int { return &b.Origin.One.Theta }),
	intRow("r 1", func(b *Base) *int { return &b.Origin.One.R }),
	intRow("φ Δ total", func(b *Base) *int { return &b.Total.Phi }),
	intRow("θ Δ total", func(b *Base) *int { return &b.Total.Theta }),
	intRow("r Δ total", func(b *Base) *int { return &b.Total.R }),
}

var sumRows = []row[Start]{
	floatRow("φ sum", func(s *Start) *float64 { return &s.Phi }),
	floatRow("θ sum", func(s *Start) *float64 { return &s.Theta }),
	floatRow("r sum", func(s *Start) *float64 { return &s.R }),
}

var nodeRows = append([]row[Col]{
	intRow("ref Δφ", func(c *Col) *int { return &c.Ref.Phi }),
	intRow("ref Δθ", func(c *Col) *int { return &c.Ref.Theta }),
	intRow("ref s", func(c *Col) *int { return &c.Ref.S }),
	floatRow("place r", func(c *Col) *float64 { return &c.At.R }),
	floatRow("place φ", func(c *Col) *float64 { return &c.At.Phi }),
	floatRow("place θ", func(c *Col) *float64 { return &c.At.Theta }),
}, append(vecRows("place", func(c *Col) *Vec3 { return &c.Pos }), vecRows("label anchor", func(c *Col) *Vec3 { return &c.Label })...)...)

var outRows = append(vecRows("edge start", func(o *OutCell) *Vec3 { return &o.Start }),
	append(vecRows("edge end", func(o *OutCell) *Vec3 { return &o.End }),
		intRow("line steps", func(o *OutCell) *int { return &o.Steps }))...)
