package Trace

type PortGeom struct {
	Name       string
	IsInput    bool
	PX, PY, PZ float64
	DX, DY, DZ float64
}

type Event struct {
	Kind     string
	Node     string
	Port     string
	Value    int
	Row, Col int
	Present  bool
	X, Y, Z  float64

	BreadcrumbLabel string
	BreadcrumbValue string
}
