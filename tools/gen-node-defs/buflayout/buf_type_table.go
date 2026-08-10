package buflayout

import "fmt"

// bufTypeInfo is the SINGLE source of truth for one buf: type tag's byte width,
// Go representation, and TS DataView accessor. Every emitter (Go writer, Go param
// type, TS getter, TS endianness flag) derives from this one table entry, so the
// Go side and TS side of a given bufType CANNOT disagree by construction — there
// used to be four parallel switches over the same bufType set (goWriterCall,
// goParamType, tsDataViewGetter, tsDataViewLE), any one of which could be edited
// without the others, producing generated files that carry the SAME
// BUF_LAYOUT_FINGERPRINT while reading DIFFERENT bytes. See
// TestBufTypeTableConsistency in main_test.go for the guard.
type bufTypeInfo struct {
	size     int                          // byte width
	goType   string                       // Go parameter type
	goWrite  func(off, val string) string // Go statement writing val into buf[off:]
	tsGetter string                       // DataView getter method name
	tsLE     bool                         // true if the getter takes a little-endian flag arg
}

// bufTypeTable is keyed by the buf: type tag string ("f32" | "i32" | "u32" | "u8").
// Add a new bufType here — and ONLY here — to teach both emitters about it.
var bufTypeTable = map[string]bufTypeInfo{
	"f32": {
		size:   4,
		goType: "float32",
		goWrite: func(off, val string) string {
			return fmt.Sprintf("binary.LittleEndian.PutUint32(buf[%s:], math.Float32bits(%s))", off, val)
		},
		tsGetter: "getFloat32",
		tsLE:     true,
	},
	"i32": {
		size:   4,
		goType: "int32",
		goWrite: func(off, val string) string {
			return fmt.Sprintf("binary.LittleEndian.PutUint32(buf[%s:], uint32(%s))", off, val)
		},
		tsGetter: "getInt32",
		tsLE:     true,
	},
	"u32": {
		size:   4,
		goType: "uint32",
		goWrite: func(off, val string) string {
			return fmt.Sprintf("binary.LittleEndian.PutUint32(buf[%s:], %s)", off, val)
		},
		tsGetter: "getUint32",
		tsLE:     true,
	},
	"u8": {
		size:   1,
		goType: "uint8",
		goWrite: func(off, val string) string {
			return fmt.Sprintf("buf[%s] = %s", off, val)
		},
		tsGetter: "getUint8",
		tsLE:     false,
	},
}

// lookupBufType returns the table entry for t, or fatalf's — a bufType missing from
// the table must be a loud build failure, not a silent default (the old switches
// each had their own silent fallback: "" from goWriterCall, "byte" from
// goParamType, "getUint8"/"" from the TS pair — all different, all wrong).
func lookupBufType(t string) bufTypeInfo {
	info, ok := bufTypeTable[t]
	if !ok {
		fatalf("unknown buf type %q (expected one of: f32, i32, u32, u8 — add it to bufTypeTable)", t)
	}
	return info
}

// bufTypeSize returns the byte width of a buf: type tag value.
func bufTypeSize(t string) (int, error) {
	info, ok := bufTypeTable[t]
	if !ok {
		return 0, fmt.Errorf("unknown buf type %q (expected f32|i32|u32|u8)", t)
	}
	return info.size, nil
}
