// event.go — the payload value types Trace's two surviving Trace-object events
// (NodeBead and Breadcrumb, trace.go) carry, plus PortGeom, the per-port geometry value
// type nodes/Wiring's own per-node stream-frame builders share. Split out of Trace.go so
// the writer type (trace.go) and the vocabulary (kind_events.go/breadcrumb_labels.go) each
// keep their own file.
package Trace

// PortGeom is one port's authoritative world geometry: its name, whether it is an
// input, its sphere-surface world position (PX/PY/PZ), and the unit direction from node
// center toward the port (DX/DY/DZ). Shared value type used by nodes/Wiring's own
// per-node stream-frame builders (node_mover.go/move_dispatch_construct.go) — independent of the
// (deleted) central NodeGeometry event this used to also ride on.
type PortGeom struct {
	Name       string
	IsInput    bool
	PX, PY, PZ float64
	DX, DY, DZ float64
}

// Event is the payload NodeBead (the one surviving Trace event) and Breadcrumb (outside
// the closed Kind vocabulary) carry. Trimmed to just the fields those two use — every
// other field the pre-decentralization Event struct carried (Step, Bead, geometry,
// camera, overlay-visibility, ...) died with the methods that populated them.
type Event struct {
	Kind     string
	Node     string
	Port     string
	Value    int
	Row, Col int
	Present  bool
	X, Y, Z  float64
	// BreadcrumbLabel/BreadcrumbValue carry a Breadcrumb() call's label/value strings.
	// Node/Port above are reused for a breadcrumb's node/port arguments.
	BreadcrumbLabel string
	BreadcrumbValue string
}
