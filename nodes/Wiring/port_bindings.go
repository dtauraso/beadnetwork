// port_bindings.go — re-exports nodes/Wiring/portwiring's PortDir/PortSpec/PortBindings
// under their historic names (god-object decomposition: the resolved per-port bindings
// data + accessors moved to their own portwiring package, since a would-be-leaf
// (PortBindings) used to hold a *MoveDispatch back-reference — see portwiring's own doc
// comment). These are plain re-exports of an already cycle-free package's types, the same
// "reuse a leaf type under its old name" shape vec_alias.go's vec3 = wire.Vec3 already
// uses — NOT an alias shim routing around a cycle: portwiring has zero dependency on
// nodes/Wiring, so there is nothing to route around. Kept because RegisterBuilder/
// PortSpec/PortBindings are part of every node kind package's public construction API
// (`Wiring.PortSpec{Name: "...", Dir: Wiring.PortIn}`) and that contract does not change
// just because the implementation moved.
package Wiring

import "github.com/dtauraso/wirefold/nodes/Wiring/portwiring"

// PortDir describes which direction a port flows.
type PortDir = portwiring.PortDir

const (
	PortIn        = portwiring.PortIn
	PortOut       = portwiring.PortOut
	PortBroadcast = portwiring.PortBroadcast
)

// PortSpec describes one port on a node kind.
type PortSpec = portwiring.PortSpec

// PortBindings holds resolved PacedWires keyed by port name — see portwiring.PortBindings.
type PortBindings = portwiring.PortBindings
