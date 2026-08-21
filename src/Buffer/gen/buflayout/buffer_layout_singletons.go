package buflayout

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

func writeBufferLayoutGoSingletons(outPath string, schema BufLayoutSchema, fp string) error {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	anyRow := false
	for _, blk := range schema.Blocks {
		if isSingletonBlock(blk.name) && hasRow(blk.name) {
			anyRow = true
			break
		}
	}
	if anyRow {
		writeBufferLayoutGoPreamble(w, fp)
	} else {
		writeBufferLayoutGoHeaderPreamble(w, fp)
	}

	for _, blk := range schema.Blocks {
		if !isSingletonBlock(blk.name) || !hasRow(blk.name) {
			continue
		}
		writeBufferLayoutGoBlockConst(w, blk)

		if blk.name == "Overlay" {
			writeRowType(w, blk)
			continue
		}

		var params []string
		for _, c := range blk.columns {
			pname := strings.ToLower(c.name[:1]) + c.name[1:]
			params = append(params, fmt.Sprintf("%s %s", pname, goParamType(c.bufType)))
		}
		fmt.Fprintf(w, "// %s writes the %s row into buf (always 1 row; no row param).\n", writerFnGoName(blk.name), blk.name)
		fmt.Fprintf(w, "func %s(buf []byte, %s) {\n", writerFnGoName(blk.name), strings.Join(params, ", "))
		for _, c := range blk.columns {
			pname := strings.ToLower(c.name[:1]) + c.name[1:]
			off := fmt.Sprintf("%d", c.offset)
			fmt.Fprintf(w, "\t%s\n", lookupBufType(c.bufType).goWrite(off, pname))
		}
		fmt.Fprintln(w, `}`)
	}

	w.Flush()
	return formatAndWrite(buf, outPath)
}

func writeRowType(w *bufio.Writer, blk bufBlock) {
	rowType := blk.name + "Row"
	fmt.Fprintf(w, "// %s is the named-field snapshot of the %s block (single row).\n", rowType, blk.name)
	fmt.Fprintf(w, "// Passed BY VALUE to %s so the write call never enumerates fields\n", writerFnGoName(blk.name))
	fmt.Fprintln(w, `// positionally — closes the swapped-adjacent-uint8-args hazard a positional`)
	fmt.Fprintln(w, `// writer call would otherwise compile silently.`)
	fmt.Fprintf(w, "type %s struct {\n", rowType)
	for _, c := range blk.columns {
		fmt.Fprintf(w, "\t%s %s\n", c.name, goParamType(c.bufType))
	}
	fmt.Fprintln(w, `}`)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "// %s writes the %s row into buf (always 1 row; no row param).\n", writerFnGoName(blk.name), blk.name)
	fmt.Fprintf(w, "func %s(buf []byte, row %s) {\n", writerFnGoName(blk.name), rowType)
	for _, c := range blk.columns {
		fname := "row." + c.name
		off := fmt.Sprintf("%d", c.offset)
		fmt.Fprintf(w, "\t%s\n", lookupBufType(c.bufType).goWrite(off, fname))
	}
	fmt.Fprintln(w, `}`)
	fmt.Fprintln(w)
}
