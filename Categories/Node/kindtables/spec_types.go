package main

type Port struct {
	ID        string
	Direction string
	Accent    string
	EdgeKind  string
	IsMulti   bool
	Optional  bool
}

type DataField struct {
	WireTag   string
	GoType    string
	FieldName string
}

type ViewDef struct {
	Kind     string
	KindID   string
	Bg       string
	Border   string
	Text     string
	MinWidth string

	Shape  string
	Fill   string
	Stroke string
	Width  string
	Height string

	Desc string
}

type KindEntry struct {
	Kind        string
	GoKind      string
	Dir         string
	KindID      uint8
	View        ViewDef
	Ports       []Port
	DataFields  []DataField
	DefaultData string
}
