// registry.go — self-registration API for node kinds.
// Each node package calls Wiring.Register in its own init().

package Wiring

// kindEntry is one entry in the kind registry.
type kindEntry struct {
	// newNode returns a fresh zero-valued pointer to the node struct.
	newNode func() any
}

// kindRegistry maps spec kind name → kindEntry. Register only records into this map;
// the loader-facing Registry (NodeBuilder, built via reflectPorts/reflectBuild) is
// populated lazily at load time by ensureRegistryBuilt (builders.go), not here — this
// keeps Register free of any dependency on the build pipeline so it can move to the
// leaf nodes/wire package.
var kindRegistry = map[string]kindEntry{}

// Register adds kind to kindRegistry. Panics if kind is already registered.
func Register(kind string, newNode func() any) {
	if _, exists := kindRegistry[kind]; exists {
		panic("Wiring.Register: kind already registered: " + kind)
	}
	kindRegistry[kind] = kindEntry{newNode: newNode}
}
