package main

// wireProp represents one wire:"prop,..." tagged field on specEdge.
type wireProp struct {
	jsonName string // from json:"..." tag
	tsType   string // from tsType:... in wire tag
	required bool   // false if "optional", true if "required"
}

// port represents one channel-typed struct field.
type port struct {
	id        string // Go field name
	direction string // "in" or "out"
	accent    string // optional hex color from SPEC.md
	edgeKind  string // optional edge kind from SPEC.md Ports table EdgeKind column
	isMulti   bool   // true when the Go type is Wiring.OutMulti
	optional  bool   // true when SPEC.md Ports table marks this port Optional=yes
}

// dataField represents a wire:"data.*" tagged struct field.
type dataField struct {
	wireTag   string // full tag value after wire:"data." prefix, e.g. "init" or "state"
	goType    string // Go type string, e.g. "[]int", "int", "string"
	fieldName string // Go struct field name (used for wire:"data.state" key derivation)
}

// viewDef holds the SPEC.md ## View section fields.
type viewDef struct {
	kind     string
	kindID   string // raw "kindId" Field/Value cell from SPEC.md's View table; "" if unassigned
	bg       string
	border   string
	text     string
	minWidth string
	// NodeTypeDef-compatible fields (used by schema/node-types consumers).
	shape  string
	fill   string
	stroke string
	width  string
	height string
	// desc is one line saying what this kind IS, from SPEC.md's "## Description" section —
	// what the palette shows under a kind's name so a person picking one does not have to
	// already know the vocabulary. Empty for a kind whose SPEC has no such section, which
	// renders as no description rather than as a placeholder.
	desc string
}

// kindEntry is one node kind to emit.
type kindEntry struct {
	kind        string // RF/view kind name (camelCase, from SPEC.md)
	goKind      string // Go/topology kind name (PascalCase, from Wiring.Register)
	dir         string // node package directory name under nodes/ (import path segment)
	kindID      uint8  // stable buffer KindId — assigned once, never renumbered by sort order
	view        viewDef
	ports       []port
	dataFields  []dataField
	defaultData string // raw JSON from SPEC.md ## Default data, or ""
}
