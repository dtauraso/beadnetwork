// Package kindscan walks nodes/<Kind>/ packages, parses each kind's Go struct
// (channel-typed port fields, wire:"data.*" tagged fields) and its SPEC.md
// (View table, Ports table, Default data block), and assembles the result
// into a KindEntry per kind. It also assigns/persists kind IDs and writes the
// kinds_generated.go blank-import list. It has no dependency on any of the
// gen-node-defs TS emitters — they consume its KindEntry output, not the
// other way around.
package kindscan

// Port represents one channel-typed struct field.
type Port struct {
	ID        string // Go field name
	Direction string // "in" or "out"
	Accent    string // optional hex color from SPEC.md
	EdgeKind  string // optional edge kind from SPEC.md Ports table EdgeKind column
	IsMulti   bool   // true when the Go type is Wiring.OutMulti
	Optional  bool   // true when SPEC.md Ports table marks this port Optional=yes
}

// DataField represents a wire:"data.*" tagged struct field.
type DataField struct {
	WireTag   string // full tag value after wire:"data." prefix, e.g. "init" or "state"
	GoType    string // Go type string, e.g. "[]int", "int", "string"
	FieldName string // Go struct field name (used for wire:"data.state" key derivation)
}

// ViewDef holds the SPEC.md ## View section fields.
type ViewDef struct {
	Kind     string
	KindID   string // raw "kindId" Field/Value cell from SPEC.md's View table; "" if unassigned
	Bg       string
	Border   string
	Text     string
	MinWidth string
	// NodeTypeDef-compatible fields (used by schema/node-types consumers).
	Shape  string
	Fill   string
	Stroke string
	Width  string
	Height string
	// Desc is one line saying what this kind IS, from SPEC.md's "## Description" section —
	// what the palette shows under a kind's name so a person picking one does not have to
	// already know the vocabulary. Empty for a kind whose SPEC has no such section, which
	// renders as no description rather than as a placeholder.
	Desc string
}

// KindEntry is one node kind to emit.
type KindEntry struct {
	Kind        string // RF/view kind name (camelCase, from SPEC.md)
	GoKind      string // Go/topology kind name (PascalCase, from Wiring.Register)
	Dir         string // node package directory name under nodes/ (import path segment)
	KindID      uint8  // stable buffer KindId — assigned once, never renumbered by sort order
	View        ViewDef
	Ports       []Port
	DataFields  []DataField
	DefaultData string // raw JSON from SPEC.md ## Default data, or ""
}
