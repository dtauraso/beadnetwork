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
	writeBufferLayoutGoPreamble(w, fp)

	for _, blk := range schema.Blocks {
		if !isSingletonBlock(blk.name) {
			continue
		}
		writeBufferLayoutGoBlockConst(w, blk)

		if blk.name == "Overlay" {
			writeOverlayRowType(w, blk)
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

func writeOverlayRowType(w *bufio.Writer, blk bufBlock) {
	fmt.Fprintln(w, `// OverlayRow is the named-field snapshot of the Overlay block (single row).`)
	fmt.Fprintln(w, `// Passed BY VALUE to SetOverlayRow so the write call never enumerates fields`)
	fmt.Fprintln(w, `// positionally — closes the swapped-adjacent-uint8-args hazard a positional`)
	fmt.Fprintln(w, `// writer call would otherwise compile silently.`)
	fmt.Fprintln(w, `type OverlayRow struct {`)
	for _, c := range blk.columns {
		fmt.Fprintf(w, "\t%s %s\n", c.name, goParamType(c.bufType))
	}
	fmt.Fprintln(w, `}`)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "// %s writes the %s row into buf (always 1 row; no row param).\n", writerFnGoName(blk.name), blk.name)
	fmt.Fprintf(w, "func %s(buf []byte, row OverlayRow) {\n", writerFnGoName(blk.name))
	for _, c := range blk.columns {
		fname := "row." + c.name
		off := fmt.Sprintf("%d", c.offset)
		fmt.Fprintf(w, "\t%s\n", lookupBufType(c.bufType).goWrite(off, fname))
	}
	fmt.Fprintln(w, `}`)
	fmt.Fprintln(w)
}
