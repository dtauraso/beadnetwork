package buflayout

import "fmt"

type bufTypeInfo struct {
	size     int
	goType   string
	goWrite  func(off, val string) string
	tsGetter string
	tsLE     bool
}

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

var runColumnTypes = map[string]bool{"bytes": true, "f32run": true}

func isRunColumn(t string) bool { return runColumnTypes[t] }

func lookupBufType(t string) bufTypeInfo {
	if isRunColumn(t) {
		fatalf("buf type %q has no row form: a run-valued column exists only as a channel, so the "+
			"block declaring it must be in movedToColumns", t)
	}
	info, ok := bufTypeTable[t]
	if !ok {
		fatalf("unknown buf type %q (expected one of: f32, i32, u32, u8, bytes, f32run — add it to bufTypeTable)", t)
	}
	return info
}

func bufTypeSize(t string) (int, error) {
	if isRunColumn(t) {
		return 0, nil
	}
	info, ok := bufTypeTable[t]
	if !ok {
		return 0, fmt.Errorf("unknown buf type %q (expected f32|i32|u32|u8|bytes|f32run)", t)
	}
	return info.size, nil
}
