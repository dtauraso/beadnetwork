// registry.go — self-registration API for node kinds.
// Each node package calls wire.Register in its own init().

package wire

// KindRegistry maps spec kind name → constructor. Register (called from each node
// package's init()) is the only writer. Read by nodes/Wiring's BuildRegistry to
// build the loader-facing NodeBuilder registry lazily at load time (reflectPorts/
// reflectBuild live in the build pipeline, not here — this keeps Register free of
// any dependency on it).
var KindRegistry = map[string]func() any{}

// Register adds kind to KindRegistry. Panics if kind is already registered.
func Register(kind string, newNode func() any) {
	if _, exists := KindRegistry[kind]; exists {
		panic("wire.Register: kind already registered: " + kind)
	}
	KindRegistry[kind] = newNode
}
